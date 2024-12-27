package config

import (
	"flag"
	"os"

	"gopkg.in/yaml.v3"
)

type CtrlConfig struct {
	// Kafka配置
	KafkaBrokers []string `yaml:"kafka_brokers"`
	KafkaTopics  []string `yaml:"kafka_topic"`
	ServerPort   string   `yaml:"server_port"`
}

var globalCtrlConfig *CtrlConfig

func defaultCtrlConfig() *CtrlConfig {
	return &CtrlConfig{
		KafkaBrokers: []string{"localhost:9092"},
		KafkaTopics:  []string{"sqcm01"},
		ServerPort:   "8088",
	}
}
func GetCtrlConfig() (*CtrlConfig, error) {
	configPath := flag.String("c", "", "Path to config file")

	flag.Parse()
	globalCtrlConfig = defaultCtrlConfig()
	data, err := os.ReadFile(*configPath)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, globalCtrlConfig); err != nil {
		return nil, err
	}
	return globalCtrlConfig, nil
}
