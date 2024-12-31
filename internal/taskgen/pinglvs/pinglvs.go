// internal/taskgen/ping_mesh/generator.go
package genpinglvs

import (
	"fmt"
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
	var tasks []models.Task

	for _, sourceNode := range g.config.Nodes {
		for _, targetNode := range g.config.Nodes {
			if sourceNode.Name == targetNode.Name {
				continue
			}

			for _, targetIP := range targetNode.LvsIPs {
				task := models.Task{
					Name:       fmt.Sprintf("%s_%s_to_%s", g.config.TaskPrefix, sourceNode.Name, targetNode.Name),
					MetricName: g.config.MetricName,
					NodeNames:  []string{sourceNode.Name},
					Interval:   g.config.Interval,
					Params: []interface{}{
						map[string]string{
							"target":     targetIP,
							"sourceNode": sourceNode.Name,
							"targetNode": targetNode.Name,
						},
					},
				}
				tasks = append(tasks, task)
			}
		}
	}

	return tasks, nil
}
