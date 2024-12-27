package storage

import (
	"fmt"
	"log"
	"strings"

	"github.com/IBM/sarama"
)

type KafkaStorage struct {
	producer sarama.AsyncProducer
	topic    string
}

func NewKafkaStorage(config *KafkaConfig) (*KafkaStorage, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.RequiredAcks = sarama.WaitForLocal
	saramaConfig.Producer.Compression = sarama.CompressionSnappy
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Producer.Return.Errors = true

	producer, err := sarama.NewAsyncProducer(config.Brokers, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka producer: %v", err)
	}

	// 处理错误和成功
	go func() {
		for err := range producer.Errors() {
			log.Printf("Failed to write to kafka: %v\n", err)
		}
	}()

	return &KafkaStorage{
		producer: producer,
		topic:    config.Topic,
	}, nil
}

func (k *KafkaStorage) Store(results []string) error {
	data := strings.Join(results, "\n")

	k.producer.Input() <- &sarama.ProducerMessage{
		Topic: k.topic,
		Value: sarama.StringEncoder(data),
	}
	return nil
}

func (k *KafkaStorage) Close() error {
	return k.producer.Close()
}
