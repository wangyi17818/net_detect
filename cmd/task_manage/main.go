// internal/task/generator.go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net_detect/internal/models"
	"time"
)

type Gateway struct {
	Name    string   `json:"name"`
	NameEn  string   `json:"name_en"`
	Gateway []string `json:"gateway"`
}

type GatewayResponse struct {
	Code int       `json:"code"`
	Data []Gateway `json:"data"`
}

func GenerateGatewayTask(nodes []string) (*models.Task, error) {
	// 获取网关数据
	resp, err := http.Get("https://cdn.monitor.just95.net/api/w1/node-gateway")
	if err != nil {
		return nil, fmt.Errorf("failed to get gateway data: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var gatewayResp GatewayResponse
	if err := json.Unmarshal(body, &gatewayResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	if gatewayResp.Code != 0 {
		return nil, fmt.Errorf("gateway api returned non-zero code: %d", gatewayResp.Code)
	}

	params := make([]interface{}, 0, len(gatewayResp.Data))
	for _, gw := range gatewayResp.Data {
		// 为每个网关创建一个任务
		for _, ip := range gw.Gateway {
			param := map[string]interface{}{
				"ip":       ip,
				"nodeName": gw.NameEn,
				"hostName": gw.Name,
			}
			params = append(params, param)
		}
	}
	task := models.Task{
		Name:       "pingMesh",
		MetricName: "pinggw",
		NodeNames:  nodes,
		Params:     params,
		Interval:   time.Minute,
	}
	return &task, nil
}

func CreateGatewayMonitorTask(controllerURL string, nodes []string) error {
	// 1. 生成任务参数
	task, err := GenerateGatewayTask(nodes)
	if err != nil {
		return fmt.Errorf("generate gateway task failed: %v", err)
	}

	// 2. 发送创建任务请求
	taskBytes, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("marshal task failed: %v", err)
	}

	resp, err := http.Post(controllerURL+"/api/tasks", "application/json", bytes.NewBuffer(taskBytes))
	if err != nil {
		return fmt.Errorf("create task request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("create task failed with status: %d", resp.StatusCode)
	}

	return nil
}

func main() {
	nodes := []string{"jhct01"}
	err := CreateGatewayMonitorTask("http://localhost:8088", nodes)
	if err != nil {
		fmt.Printf("Failed to create gateway monitor task: %v\n", err)
	}
}
