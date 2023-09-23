/* Package utils provides utility functions for the cnexporter project.*/
package utils

import (
	"strings"

	"github.com/docker/docker/api/types"
)

type CntLabels struct {
	ID     string
	Image  string
	Name   string
	Status string
	State  string
}

func CountByStatus(containers []types.Container, state string) int {
	count := 0
	for _, cnt := range containers {
		if cnt.State == state {
			count++
		}
	}
	return count
}

func BuildLabels(container types.Container) CntLabels {
	return CntLabels{
		container.ID,
		container.Image,
		strings.Trim(container.Names[0], "/"),
		container.Status,
		container.State,
	}
}
