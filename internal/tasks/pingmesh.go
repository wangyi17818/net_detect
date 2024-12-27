package tasks

import (
	"encoding/json"
	"fmt"
	"net_detect/internal/models"
	"net_detect/internal/ping"
	"net_detect/internal/storage"
	"net_detect/utils"
	"strings"
	"sync"
)

// PingMeshTask pingMesh任务实现
type PingMeshTask struct {
	pinger     ping.Pinger
	storage    storage.ResultStorage
	metricName string
}

func NewPingMeshTask(pinger ping.Pinger, storage storage.ResultStorage) *PingMeshTask {
	return &PingMeshTask{
		pinger:  pinger,
		storage: storage,
	}
}

func (t *PingMeshTask) Name() string {
	return "pingMesh"
}

func (t *PingMeshTask) Execute(metricName string, params []interface{}) error {
	targets, err := t.parseParams(params)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	results := make([]models.PingResult, len(targets))

	// 并发执行ping
	for i, target := range targets {
		wg.Add(1)
		go func(index int, target models.PingTarget) {
			defer wg.Done()
			results[index] = t.pinger.Ping(target)
		}(i, target)
	}
	wg.Wait()

	return t.storage.Store(t.resultInfulxDBFormat(metricName, results))
}

func (t *PingMeshTask) resultInfulxDBFormat(metricName string, results []models.PingResult) []string {
	// 构建Influx行协议数据
	var lines []string
	source_nodename, _ := utils.GetNodeName()
	source_hostname, _ := utils.GetHostName()
	for _, r := range results {
		// 合并默认tags和自定义tags
		tags := make([]string, 0)
		tags = append(tags,
			fmt.Sprintf("source_ip=%s", r.SourceIP),
			fmt.Sprintf("source_node=%s", source_nodename),
			fmt.Sprintf("source_host=%s", source_hostname),
			fmt.Sprintf("target_ip=%s", r.TargetIP),
			fmt.Sprintf("target_node=%s", r.TargetNode),
			fmt.Sprintf("target_host=%s", r.TargetHost),
			fmt.Sprintf("ip_version=%s", r.IPVersion),
		)
		for k, v := range r.Tags {
			tags = append(tags, fmt.Sprintf("%s=%s", k, v))
		}

		// 构建fields
		fields := []string{
			fmt.Sprintf("packets_sent=%di", r.PacketsSent),
			fmt.Sprintf("packets_recv=%di", r.PacketsRecv),
			fmt.Sprintf("packets_loss=%di", r.PacketsLoss),
			fmt.Sprintf("rtt_min=%f", r.MinRtt),
			fmt.Sprintf("rtt_max=%f", r.MaxRtt),
			fmt.Sprintf("rtt_avg=%f", r.AvgRtt),
			fmt.Sprintf("rtt_std_dev=%f", r.StdDevRtt),
		}
		if r.Error != "" {
			fields = append(fields, fmt.Sprintf("error=%v", r.Error))
		}

		// 构建完整的行
		line := fmt.Sprintf("%s,%s %s %d",
			metricName,
			strings.Join(tags, ","),
			strings.Join(fields, ","),
			r.Timestamp.UnixNano(),
		)
		lines = append(lines, line)
	}
	return lines
}

func (t *PingMeshTask) parseParams(params []interface{}) ([]models.PingTarget, error) {
	targets := make([]models.PingTarget, 0, len(params))
	for _, param := range params {
		paramMap, ok := param.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid parameter type: expected map[string]interface{}, got %T", param)
		}

		// 将 map 转换为 JSON
		paramJSON, err := json.Marshal(paramMap)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal parameter: %v", err)
		}

		// 将 JSON 解析为 PingTarget
		var target models.PingTarget
		if err := json.Unmarshal(paramJSON, &target); err != nil {
			return nil, fmt.Errorf("failed to unmarshal parameter to PingTarget: %v", err)
		}

		// 验证必要字段
		if target.IP == "" {
			return nil, fmt.Errorf("IP field is required")
		}

		targets = append(targets, target)
	}

	if len(targets) == 0 {
		return nil, fmt.Errorf("no valid targets found in parameters")
	}

	return targets, nil
}
