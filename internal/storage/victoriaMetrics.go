package storage

import (
	"bytes"
	"fmt"
	"net/http"
	"net_detect/internal/config"
	"strings"
)

type VictoriaMetricsStorage struct {
	client *http.Client
	config *VictoriaMetricsConfig
}

func NewVictoriaMetricsStorage(config *VictoriaMetricsConfig) (*VictoriaMetricsStorage, error) {
	return &VictoriaMetricsStorage{
		client: &http.Client{
			Timeout: config.Timeout,
		},
		config: config,
	}, nil
}

func (v *VictoriaMetricsStorage) Store(results []string) error {
	// 发送数据
	data := strings.Join(results, "\n")
	req, err := http.NewRequest("POST", v.config.Address+"/insert/0/influx/write", bytes.NewBufferString(data))
	if err != nil {
		return fmt.Errorf("create request failed: %v", err)
	}

	req.SetBasicAuth(v.config.Username, v.config.Password)
	req.Header.Set("Content-Type", "text/plain")
	resp, err := v.client.Do(req)

	globalConfig := config.Get()
	for i := 0; i < globalConfig.VMMaxRetries; i++ {
		if err != nil {
			fmt.Printf("send request failed num: %v", err)
		} else {
			break
		}
		resp, err = v.client.Do(req)
	}
	if err != nil {
		return fmt.Errorf("send request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		// TODO: 添加重试机制，使用globalConfig.VMAddress 中的其它地址
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (v *VictoriaMetricsStorage) Close() error {
	return nil
}
