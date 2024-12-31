// internal/taskgen/gateway_ping/config.go
package gatewayping

import "time"

type Config struct {
	GatewayAPIURL string
	Nodes         []string
	Interval      time.Duration
	MetricName    string
	TaskPrefix    string
}

type Gateway struct {
	Name    string   `json:"name"`
	NameEn  string   `json:"name_en"`
	Gateway []string `json:"gateway"`
}

type GatewayResponse struct {
	Code int       `json:"code"`
	Data []Gateway `json:"data"`
}
