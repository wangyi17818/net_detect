// cmd/controller/main.go
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"net_detect/internal/api"
	"net_detect/internal/config"
	"net_detect/internal/controller"
)

// cmd/controller/main.go
func main() {
	conf, err := config.GetCtrlConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 创建控制器
	ctrl, err := controller.NewController(
		conf.KafkaBrokers,
	)
	if err != nil {
		log.Fatalf("Failed to create controller: %v", err)
	}

	// 创建并启动 API 服务
	server := api.NewServer(ctrl)
	go func() {
		if err := server.Start(":" + conf.ServerPort); err != nil {
			log.Fatalf("Failed to start API server: %v", err)
		}
	}()

	// 启动控制器
	ctrl.Start()

	// 等待中断信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	// 优雅退出
	ctrl.Stop()
}
