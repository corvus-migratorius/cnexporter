package exporter

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/prometheus/client_golang/prometheus"
)

type CntLabels struct {
	Id     string
	Image  string
	Name   string
	Status string
	State  string
}

type CntCounts struct {
	Total   *prometheus.GaugeVec
	Created *prometheus.GaugeVec
	Running *prometheus.GaugeVec
	Exited  *prometheus.GaugeVec
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

func RecordCounts(counts CntCounts, dclient *client.Client, dcontext context.Context) {
	go func() {
		for {
			containers, err := dclient.ContainerList(dcontext, types.ContainerListOptions{All: true})
			if err != nil {
				panic(err)
			}

			hostname, err := os.Hostname()
			if err != nil {
				log.Fatal("Failed to read the hostname")
			}

			cnt_total := float64(len(containers))
			cnt_created := float64(CountByStatus(containers, "created"))
			cnt_running := float64(CountByStatus(containers, "running"))
			cnt_exited := float64(CountByStatus(containers, "exited"))

			counts.Total.With(prometheus.Labels{"nodename": hostname}).Set(cnt_total)
			counts.Created.With(prometheus.Labels{"nodename": hostname}).Set(cnt_created)
			counts.Running.With(prometheus.Labels{"nodename": hostname}).Set(cnt_running)
			counts.Exited.With(prometheus.Labels{"nodename": hostname}).Set(cnt_exited)

			time.Sleep(15 * time.Second)
		}
	}()
}

func RecordMetadata(metadata *prometheus.GaugeVec, dclient *client.Client, dcontext context.Context) {
	go func() {
		for {
			containers, err := dclient.ContainerList(dcontext, types.ContainerListOptions{All: true})
			if err != nil {
				panic(err)
			}

			hostname, err := os.Hostname()
			if err != nil {
				log.Fatal("Failed to read the hostname")
			}

			metadata.Reset()
			for _, cnt := range containers {
				labels := BuildLabels(cnt)
				metadata.With(prometheus.Labels{
					"id":       labels.Id,
					"image":    labels.Image,
					"name":     labels.Name,
					"status":   labels.Status,
					"state":    labels.State,
					"nodename": hostname,
				})
			}

			time.Sleep(15 * time.Second)
		}
	}()
}
