package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"net_detect/internal/agent"
	"net_detect/internal/config"
	"net_detect/internal/ping"
	"net_detect/internal/storage"
	"net_detect/internal/tasks"
	"net_detect/utils"
)

func main() {
	conf, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	nodeName, _ := utils.GetNodeName()
	//hostName, _ := utils.GetHostName()
	// 创建存储
	var storageConfig storage.Config
	if conf.StorageType == "kafka" {
		storageConfig = storage.Config{
			Type: storage.StorageTypeKafka,
			KafkaConfig: &storage.KafkaConfig{
				Brokers:      conf.KafkaBrokers,
				Topic:        "net_detect_result",
				BatchSize:    100,
				BatchTimeout: 5 * time.Second,
			},
		}
	} else {
		storageConfig = storage.Config{
			Type: storage.StorageTypeVictoriaMetrics,
			VictoriaMetricsConfig: &storage.VictoriaMetricsConfig{
				Address:  conf.VMAddress[0],
				Username: conf.VMUsername,
				Password: conf.VMPassword,
				Timeout:  10 * time.Second,
			},
		}
	}

	resultStorage, err := storage.NewStorage(storageConfig)
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}
	defer resultStorage.Close()

	// 创建Agent配置
	agentConfig := agent.Config{
		KafkaBrokers: conf.KafkaBrokers,
		KafkaGroup:   fmt.Sprintf("%s-agent", nodeName),
		KafkaTopic:   fmt.Sprintf("%s-task", nodeName),
	}
	log.Printf("Topic: %+v", agentConfig.KafkaTopic)

	// 创建Agent
	agent, err := agent.NewAgent(agentConfig)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	// 创建ping执行器和任务
	pinger := ping.NewPinger(ping.DefaultConfig())
	pingMeshTask := tasks.NewPingMeshTask(pinger, resultStorage)

	// 注册任务
	agent.RegisterTask(pingMeshTask)

	// 启动Agent
	go func() {
		if err := agent.Start(); err != nil {
			log.Printf("Agent error: %v", err)
		}
	}()

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// 优雅退出
	agent.Stop()
}
