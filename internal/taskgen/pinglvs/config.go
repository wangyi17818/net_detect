// internal/taskgen/ping_mesh/config.go
package genpinglvs

import "time"

type Node struct {
	Name   string
	LvsIPs []string
}

type Config struct {
	Nodes      []Node
	Interval   time.Duration
	MetricName string
	TaskPrefix string
}
