package tasks

// Task 定义任务接口
type Task interface {
	Name() string
	Execute(metricName string, params []any) error
}
