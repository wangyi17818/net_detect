package storage

import (
	"fmt"
	"time"
)

// ResultStorage 结果存储接口
type ResultStorage interface {
	Store(results []string) error
	Close() error
}

// StorageType 存储类型
type StorageType string

const (
	StorageTypeKafka           StorageType = "kafka"
	StorageTypeVictoriaMetrics StorageType = "victoriametrics"
)

// Config 存储配置
type Config struct {
	Type StorageType
	// Kafka配置
	KafkaConfig *KafkaConfig
	// VictoriaMetrics配置
	VictoriaMetricsConfig *VictoriaMetricsConfig
}

// KafkaConfig Kafka配置
type KafkaConfig struct {
	Brokers      []string
	Topic        string
	BatchSize    int
	BatchTimeout time.Duration
}

// VictoriaMetricsConfig VictoriaMetrics配置
type VictoriaMetricsConfig struct {
	Address  string
	Username string
	Password string
	Timeout  time.Duration
}

// NewStorage 创建存储实例
func NewStorage(config Config) (ResultStorage, error) {
	switch config.Type {
	case StorageTypeKafka:
		return NewKafkaStorage(config.KafkaConfig)
	case StorageTypeVictoriaMetrics:
		return NewVictoriaMetricsStorage(config.VictoriaMetricsConfig)
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", config.Type)
	}
}
