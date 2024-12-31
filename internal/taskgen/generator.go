// internal/taskgen/generator.go
package taskgen

import "net_detect/internal/models"

type Generator interface {
	Generate() ([]models.Task, error)
}
