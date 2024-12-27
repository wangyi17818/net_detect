package models

import (
	"fmt"
	"time"
)

// TaskMessage Kafka任务消息
type TaskMessage struct {
	TaskName   string        `json:"taskName"`
	MetricName string        `json:"metricName"`
	Params     []interface{} `json:"params"`
}

// PingTarget 探测目标
type PingTarget struct {
	IP       string            `json:"ip"`
	NodeName string            `json:"nodeName"`
	HostName string            `json:"hostName"`
	Tags     map[string]string `json:"tags,omitempty"` // 附加标签
}

// PingResult 探测结果
type PingResult struct {
	SourceIP    string            `json:"sourceIp"`
	TargetIP    string            `json:"targetIp"`
	TargetNode  string            `json:"targetNode"`
	TargetHost  string            `json:"targetHost"`
	Tags        map[string]string `json:"tags,omitempty"`
	PacketsSent int               `json:"packetsSent"`
	PacketsRecv int               `json:"packetsRecv"`
	PacketsLoss int               `json:"packetsLoss"`
	MinRtt      float64           `json:"minRtt"`
	MaxRtt      float64           `json:"maxRtt"`
	AvgRtt      float64           `json:"avgRtt"`
	StdDevRtt   float64           `json:"stdDevRtt"`
	IPVersion   string            `json:"ipVersion"`
	Error       string            `json:"error,omitempty"`
	Timestamp   time.Time         `json:"timestamp"`
}

// controller 控制器接口

// Task 定义一个任务
type Task struct {
	Name       string            `json:"name"`       // 任务名称，如 pingMesh
	MetricName string            `json:"metricName"` // 任务ID
	NodeNames  []string          `json:"nodeNames"`  // 执行任务的节点列表
	Params     []interface{}     `json:"params"`     // 任务参数
	Interval   time.Duration     `json:"interval"`   // 执行频率
	Tags       map[string]string `json:"tags"`       // 任务标签，可选
}

// GetTopics 获取任务对应的所有topics
func (t *Task) GetTopics() []string {
	topics := make([]string, len(t.NodeNames))
	for i, node := range t.NodeNames {
		// 生成格式：task-{nodeName}
		topics[i] = fmt.Sprintf("%s-task", node)
	}
	return topics
}
