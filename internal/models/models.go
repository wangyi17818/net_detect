package models

// BaseResult 基础结果结构
type BaseResult struct {
	TaskName  string `json:"taskName"`
	Timestamp int64  `json:"timestamp"`
}
