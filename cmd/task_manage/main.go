// cmd/task_manage/main.go
package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"net_detect/internal/controller"
	gatewayping "net_detect/internal/taskgen/pinggw"
)

type Config struct {
	KafkaBrokers  []string
	GatewayAPIURL string
	Nodes         []string
	Interval      time.Duration
}

func main() {
	// 1. 解析配置
	var config Config
	kafkaBrokers := flag.String("brokers", "localhost:9092", "Kafka brokers")
	gatewayURL := flag.String("gateway-url", "https://cdn.monitor.just95.net/api/w1/node-gateway", "Gateway API URL")
	interval := flag.Duration("interval", 5*time.Minute, "Task interval")
	flag.Parse()

	config.KafkaBrokers = strings.Split(*kafkaBrokers, ",")
	config.GatewayAPIURL = *gatewayURL
	config.Interval = *interval

	// 2. 创建controller
	ctrl, err := controller.NewController(config.KafkaBrokers)
	if err != nil {
		log.Fatalf("Failed to create controller: %v", err)
	}
	defer ctrl.Stop()

	// 3. 创建Gateway Ping任务生成器
	gwGenerator := gatewayping.NewGenerator(gatewayping.Config{
		GatewayAPIURL: config.GatewayAPIURL,
		Nodes:         config.Nodes,
		Interval:      config.Interval,
		MetricName:    "gateway_ping",
		TaskPrefix:    "gw_ping",
	})
	// No need to stop gwGenerator as it does not have a Stop method

	// 4. 生成并添加任务
	tasks, err := gwGenerator.Generate()
	if err != nil {
		log.Printf("Failed to generate gateway tasks: %v", err)
	}

	for _, task := range tasks {
		if err := ctrl.AddTask(task); err != nil {
			log.Printf("Failed to add task %s: %v", task.Name, err)
		}
	}

	// 5. 保持运行
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
}
