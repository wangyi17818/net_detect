// internal/taskgen/gateway_ping/generator.go
package gatewayping

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net_detect/internal/models"
	"net_detect/internal/taskgen"
)

type Generator struct {
	config Config
}

func NewGenerator(config Config) taskgen.Generator {
	return &Generator{config: config}
}

func (g *Generator) Generate() ([]models.Task, error) {
	// 获取网关数据
	resp, err := http.Get(g.config.GatewayAPIURL)
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

	var tasks []models.Task

	// 为每个节点创建ping网关的任务
	for _, node := range g.config.Nodes {
		task := models.Task{
			Name:       fmt.Sprintf("%s_%s", g.config.TaskPrefix, node),
			MetricName: g.config.MetricName,
			NodeNames:  []string{node},
			Interval:   g.config.Interval,
			Params:     g.generateParams(gatewayResp.Data),
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (g *Generator) generateParams(gateways []Gateway) []interface{} {
	// 提取所有网关IP
	var targets []string
	for _, gw := range gateways {
		targets = append(targets, gw.Gateway...)
	}

	var params []interface{}
	for _, target := range targets {
		params = append(params, target)
	}
	return params
}
