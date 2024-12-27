package ping

import (
	"fmt"
	"time"

	"net_detect/internal/config"
	"net_detect/internal/models"
	"net_detect/utils"

	probing "github.com/prometheus-community/pro-bing"
)

// Config ping配置
type Config struct {
	Count    int
	Interval time.Duration
	Timeout  time.Duration
}

// DefaultConfig 默认配置
func DefaultConfig() Config {
	conf := config.Get()
	return Config{
		Count:    conf.PingCount,
		Interval: conf.PingInterval,
		Timeout:  conf.PingTimeout,
	}
}

// Pinger ping执行器接口
type Pinger interface {
	Ping(target models.PingTarget) models.PingResult
}

// DefaultPinger 默认的ping实现
type DefaultPinger struct {
	config Config
}

func NewPinger(config Config) *DefaultPinger {
	return &DefaultPinger{config: config}
}

func (p *DefaultPinger) Ping(target models.PingTarget) models.PingResult {
	pinger, err := probing.NewPinger(target.IP)
	if err != nil {
		return models.PingResult{
			TargetIP:   target.IP,
			TargetNode: target.NodeName,
			TargetHost: target.HostName,
			Tags:       target.Tags,
			Error:      fmt.Sprintf("创建pinger失败: %v", err),
			Timestamp:  time.Now(),
		}
	}

	pinger.Count = p.config.Count
	pinger.Interval = p.config.Interval
	pinger.Timeout = p.config.Timeout
	pinger.SetPrivileged(true)

	err = pinger.Run()
	if err != nil {
		return models.PingResult{
			SourceIP:   pinger.Source,
			TargetIP:   target.IP,
			TargetNode: target.NodeName,
			TargetHost: target.HostName,
			Tags:       target.Tags,
			Error:      fmt.Sprintf("执行ping失败: %v", err),
			Timestamp:  time.Now(),
			IPVersion:  utils.GetIPVersion(target.IP),
		}
	}
	source_ip := pinger.Source
	if source_ip == "" {
		source_ip, _ = utils.GetLocalIP(target.IP)
	}
	stats := pinger.Statistics()
	return models.PingResult{
		SourceIP:    source_ip,
		TargetIP:    target.IP,
		TargetNode:  target.NodeName,
		TargetHost:  target.HostName,
		Tags:        target.Tags,
		PacketsSent: stats.PacketsSent,
		PacketsRecv: stats.PacketsRecv,
		PacketsLoss: stats.PacketsSent - stats.PacketsRecv,
		MinRtt:      float64(stats.MinRtt) / float64(time.Millisecond),
		MaxRtt:      float64(stats.MaxRtt) / float64(time.Millisecond),
		AvgRtt:      float64(stats.AvgRtt) / float64(time.Millisecond),
		StdDevRtt:   float64(stats.StdDevRtt) / float64(time.Millisecond),
		IPVersion:   utils.GetIPVersion(target.IP),
		Timestamp:   time.Now(),
	}
}
