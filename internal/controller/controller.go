// internal/controller/controller.go
package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"net_detect/internal/models"

	"github.com/IBM/sarama"
)

type Controller struct {
	producer  sarama.SyncProducer
	tasks     map[string]models.Task
	runners   map[string]chan struct{}
	stopCh    chan struct{}
	taskMutex sync.RWMutex
}

func NewController(brokers []string) (*Controller, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %v", err)
	}

	return &Controller{
		producer: producer,
		tasks:    make(map[string]models.Task),
		runners:  make(map[string]chan struct{}),
		stopCh:   make(chan struct{}),
	}, nil
}

// ListTasks 获取所有任务
func (c *Controller) ListTasks() []models.Task {
	c.taskMutex.RLock()
	defer c.taskMutex.RUnlock()

	tasks := make([]models.Task, 0, len(c.tasks))
	for _, t := range c.tasks {
		tasks = append(tasks, t)
	}
	return tasks
}

// GetTask 获取单个任务
func (c *Controller) GetTask(taskID string) (models.Task, bool) {
	c.taskMutex.RLock()
	defer c.taskMutex.RUnlock()

	t, exists := c.tasks[taskID]
	return t, exists
}

// AddTask 添加任务
func (c *Controller) AddTask(t models.Task) error {
	c.taskMutex.Lock()
	defer c.taskMutex.Unlock()

	if _, exists := c.tasks[t.MetricName]; exists {
		close(c.runners[t.MetricName])
	}

	c.tasks[t.MetricName] = t
	stopCh := make(chan struct{})
	c.runners[t.MetricName] = stopCh
	go c.runTask(t, stopCh)
	return nil
}

// ReloadTask 重写任务
func (c *Controller) AppendTask(t models.Task) error {
	c.taskMutex.Lock()
	defer c.taskMutex.Unlock()

	if _, exists := c.tasks[t.MetricName]; exists {
		close(c.runners[t.MetricName])
	}

	c.tasks[t.MetricName] = t
	stopCh := make(chan struct{})
	c.runners[t.MetricName] = stopCh
	go c.runTask(t, stopCh)
	return nil
}

// RemoveTask 移除任务
func (c *Controller) RemoveTask(taskID string) error {
	c.taskMutex.Lock()
	defer c.taskMutex.Unlock()

	if stopCh, exists := c.runners[taskID]; exists {
		close(stopCh)
		delete(c.runners, taskID)
		delete(c.tasks, taskID)
		return nil
	}
	return fmt.Errorf("task %s not found", taskID)
}

func (c *Controller) Stop() {
	close(c.stopCh)
	c.taskMutex.Lock()
	for _, stopCh := range c.runners {
		close(stopCh)
	}
	c.taskMutex.Unlock()

	if err := c.producer.Close(); err != nil {
		log.Printf("Failed to close producer: %v", err)
	}
}

// sendTaskMessages 为每个节点发送任务消息到对应的topic
func (c *Controller) sendTaskMessages(t models.Task) error {
	// 获取任务的所有topics
	topics := t.GetTopics()

	for i, topic := range topics {
		msg := models.TaskMessage{
			TaskName:   t.Name,
			MetricName: t.MetricName,
			Params:     t.Params,
		}

		msgBytes, err := json.Marshal(msg)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %v", err)
		}

		// 发送到对应节点的topic
		_, _, err = c.producer.SendMessage(&sarama.ProducerMessage{
			Topic: topic,
			Value: sarama.StringEncoder(msgBytes),
		})
		if err != nil {
			log.Printf("Failed to send message to topic %s: %v", topic, err)
			continue
		}

		log.Printf("Sent task %s to node %s via topic %s, len: %v", t.MetricName, t.NodeNames[i], topic, len(t.Params))
	}
	return nil
}

// Start 启动控制器
func (c *Controller) Start() error {
	c.taskMutex.RLock()
	defer c.taskMutex.RUnlock()

	// 检查 producer 是否就绪
	if c.producer == nil {
		return fmt.Errorf("producer is not initialized")
	}

	// 遍历所有已存在的任务并启动
	for taskID, task := range c.tasks {
		if _, exists := c.runners[taskID]; !exists {
			// 为每个任务创建一个停止通道
			stopCh := make(chan struct{})
			c.runners[taskID] = stopCh
			go c.runTask(task, stopCh)
			log.Printf("Started task: %s", taskID)
		}
	}

	log.Printf("Controller started with %d tasks", len(c.tasks))
	return nil
}

// runTask 运行单个任务
func (c *Controller) runTask(t models.Task, stopCh chan struct{}) {
	ticker := time.NewTicker(t.Interval)
	defer ticker.Stop()

	// 立即执行一次
	if err := c.sendTaskMessages(t); err != nil {
		log.Printf("Failed to send initial task messages for task %s: %v", t.MetricName, err)
	}

	for {
		select {
		case <-c.stopCh:
			log.Printf("Stopping task %s due to controller shutdown", t.MetricName)
			return
		case <-stopCh:
			log.Printf("Stopping task %s", t.MetricName)
			return
		case <-ticker.C:
			if err := c.sendTaskMessages(t); err != nil {
				log.Printf("Failed to send task messages for task %s: %v", t.MetricName, err)
			}
		}
	}
}
