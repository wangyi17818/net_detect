// internal/config/config.go
package config

import (
	"flag"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	// Kafka配置
	KafkaBrokers     []string `yaml:"kafka_brokers"`
	KafkaGroup       string   `yaml:"kafka_group"`
	KafkaTopic       string   `yaml:"kafka_topic"`
	KafkaResultTopic string   `yaml:"kafka_result_topic"`

	// VictoriaMetrics配置
	VMAddress    []string      `yaml:"vm_address"`
	VMUsername   string        `yaml:"vm_username"`
	VMPassword   string        `yaml:"vm_password"`
	VMTimeout    time.Duration `yaml:"vm_timeout"`
	VMMaxRetries int           `yaml:"vm_max_retries"`

	// Ping配置
	PingCount    int           `yaml:"ping_count"`
	PingInterval time.Duration `yaml:"ping_interval"`
	PingTimeout  time.Duration `yaml:"ping_timeout"`

	// 存储类型
	StorageType string `yaml:"storage_type"`
}

var globalConfig *Config

// 默认配置
func defaultConfig() *Config {
	return &Config{
		KafkaBrokers:     []string{"localhost:9092"},
		KafkaGroup:       "ping-agent",
		KafkaTopic:       "ping-tasks",
		KafkaResultTopic: "netdetect-results",
		VMAddress:        []string{"localhost:8428"},
		VMUsername:       "net_detect",
		VMPassword:       "",
		VMTimeout:        10 * time.Second,
		VMMaxRetries:     3,
		PingCount:        10,
		PingInterval:     100 * time.Millisecond,
		PingTimeout:      1000 * time.Millisecond,
		StorageType:      "victoriametrics",
	}
}

// 获取全局配置
func Get() *Config {
	return globalConfig
}

func LoadConfig() (*Config, error) {
	// 加载默认配置
	globalConfig = defaultConfig()

	// 命令行参数
	configPath := flag.String("c", "", "Path to config file")
	kafkaBrokers := flag.String("kafka-brokers", "", "Kafka brokers")
	kafkaGroup := flag.String("kafka-group", "", "Kafka consumer group")
	kafkaTopic := flag.String("kafka-topic", "", "Kafka topic for tasks")
	kafkaResultTopic := flag.String("kafka-result-topic", "netdetect-results", "Kafka topic for results")

	vmAddr := flag.String("vm-addr", "", "VictoriaMetrics address")
	vmUser := flag.String("vm-user", "", "VictoriaMetrics username")
	vmPass := flag.String("vm-pass", "", "VictoriaMetrics password")
	vmTimeout := flag.Duration("vm-timeout", 0, "VictoriaMetrics timeout")
	vmRetries := flag.Int("vm-retries", 0, "VictoriaMetrics max retries")
	storageType := flag.String("storage", "", "Storage type: kafka or victoriametrics")
	// ping参数
	pingCount := flag.Int("ping-count", 0, "Number of ping packets to send")
	pingInterval := flag.Duration("ping-interval", 0, "Interval between ping packets")
	pingTimeout := flag.Duration("ping-timeout", 0, "Ping timeout")

	flag.Parse()
	// 如果指定了配置文件，则读取配置文件
	if *configPath != "" {
		data, err := os.ReadFile(*configPath)
		if err != nil {
			return nil, err
		}

		if err := yaml.Unmarshal(data, globalConfig); err != nil {
			return nil, err
		}
	}

	// 命令行参数优先
	if *kafkaBrokers != "" {
		globalConfig.KafkaBrokers = strings.Split(*kafkaBrokers, ",")
	}
	if *kafkaGroup != "" {
		globalConfig.KafkaGroup = *kafkaGroup
	}
	if *kafkaTopic != "" {
		globalConfig.KafkaTopic = *kafkaTopic
	}
	if *kafkaResultTopic != "" {
		globalConfig.KafkaResultTopic = *kafkaResultTopic
	}
	if *vmAddr != "" {
		globalConfig.VMAddress = strings.Split(*vmAddr, ",")
	}
	if *vmUser != "" {
		globalConfig.VMUsername = *vmUser
	}
	if *vmPass != "" {
		globalConfig.VMPassword = *vmPass
	}
	if *vmTimeout != 0 {
		globalConfig.VMTimeout = *vmTimeout
	}
	if *vmRetries != 0 {
		globalConfig.VMMaxRetries = *vmRetries
	}
	if *storageType != "" {
		globalConfig.StorageType = *storageType
	}
	if *pingCount != 0 {
		globalConfig.PingCount = *pingCount
	}
	if *pingInterval != 0 {
		globalConfig.PingInterval = *pingInterval
	}
	if *pingTimeout != 0 {
		globalConfig.PingTimeout = *pingTimeout
	}

	return globalConfig, nil
}
