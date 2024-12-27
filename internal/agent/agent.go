package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"net_detect/internal/models"
	"net_detect/internal/tasks"

	"github.com/IBM/sarama"
)

type Config struct {
	KafkaBrokers []string
	KafkaGroup   string
	KafkaTopic   string
}

// internal/agent/agent.go

type Agent struct {
	consumer sarama.ConsumerGroup
	tasks    map[string]tasks.Task
	ctx      context.Context
	cancel   context.CancelFunc
	config   Config // 添加 config 字段
}

func NewAgent(config Config) (*Agent, error) {
	// Kafka消费者配置
	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	kafkaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	kafkaConfig.Consumer.Group.Session.Timeout = 20 * time.Second
	kafkaConfig.Consumer.Group.Heartbeat.Interval = 6 * time.Second

	// 创建消费者组
	consumer, err := sarama.NewConsumerGroup(config.KafkaBrokers, config.KafkaGroup, kafkaConfig)
	if err != nil {
		return nil, fmt.Errorf("create consumer group failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Agent{
		consumer: consumer,
		tasks:    make(map[string]tasks.Task),
		ctx:      ctx,
		cancel:   cancel,
		config:   config, // 保存配置
	}, nil
}

func (a *Agent) RegisterTask(task tasks.Task) {
	a.tasks[task.Name()] = task
}

func (a *Agent) Start() error {
	handler := &ConsumerGroupHandler{agent: a}
	for {
		select {
		case <-a.ctx.Done():
			return nil
		default:
			if err := a.consumer.Consume(a.ctx, []string{a.config.KafkaTopic}, handler); err != nil {
				log.Printf("Error from consumer: %v", err)
			}
		}
	}
}

func (a *Agent) Stop() {
	a.cancel()
	if err := a.consumer.Close(); err != nil {
		log.Printf("Error closing consumer: %v", err)
	}
}

type ConsumerGroupHandler struct {
	agent *Agent
}

func (h *ConsumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *ConsumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		var task models.TaskMessage
		if err := json.Unmarshal(message.Value, &task); err != nil {
			log.Printf("Failed to unmarshal task: %v", err)
			continue
		}
		log.Printf("Received task: %+v, task targets size: %v", task.TaskName, len(task.Params))

		handler, exists := h.agent.tasks[task.TaskName]
		if !exists {
			log.Printf("Unknown task type: %s", task.TaskName)
			continue
		}
		if task.MetricName == "" {
			task.MetricName = task.TaskName
		}
		if err := handler.Execute(task.MetricName, task.Params); err != nil {
			log.Printf("Failed to execute task %s: %v", task.TaskName, err)
		}
		log.Printf("Task: %v, finished", task.TaskName)

		session.MarkMessage(message, "")
	}
	return nil
}
