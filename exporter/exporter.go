// Package cnexporter exports data about Docker containers as Prometheus gauges.
//
// Data extraction is handled with the `containerExporter` type.
// Use the `ContainerExporter` factory function to get a properly initialized
// instance of the exporter.
//
// A web server exporting the `/metrics` endpoint is not included but is trivial
// to implement.

package exporter

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"docker-exporter/utils"
)

// Unifies four Gauge vectors in a single struct for convenience
type CntCounts struct {
	Total   *prometheus.GaugeVec
	Created *prometheus.GaugeVec
	Running *prometheus.GaugeVec
	Exited  *prometheus.GaugeVec
}

// Factory function returning pointers to properly initialized containerExporter instances
func ContainerExporter(context context.Context, client *client.Client, timeout int) *containerExporter {
	exporter := containerExporter{
		Context: context,
		Client:  client,
		Timeout: timeout,
	}

	exporter.initCountGauges()
	exporter.initMetadataGauge()
	return &exporter
}

// Creates, registeres and updates Prometheus GaugeVecs in goroutines
//
// Counts: predefined GaugeVecs for "total", "created", "running" and "exited" container counts.
// Metadata: always reports 0 but holds container metadata as labels.
type containerExporter struct {
	Context  context.Context
	Client   *client.Client
	Counts   CntCounts
	Metadata *prometheus.GaugeVec
	Timeout  int
}

// Update container count metrics in a goroutine every `timeout` seconds
func (self *containerExporter) RecordCounts() {
	go func() {
		for {
			containers := self.getContainers()
			hostname := self.getHostname()

			cnt_total := float64(len(containers))
			cnt_created := float64(utils.CountByStatus(containers, "created"))
			cnt_running := float64(utils.CountByStatus(containers, "running"))
			cnt_exited := float64(utils.CountByStatus(containers, "exited"))

			self.Counts.Total.With(prometheus.Labels{"nodename": hostname}).Set(cnt_total)
			self.Counts.Created.With(prometheus.Labels{"nodename": hostname}).Set(cnt_created)
			self.Counts.Running.With(prometheus.Labels{"nodename": hostname}).Set(cnt_running)
			self.Counts.Exited.With(prometheus.Labels{"nodename": hostname}).Set(cnt_exited)

			time.Sleep(time.Duration(self.Timeout) * time.Second)
		}
	}()
}

// Update container metadata metrics in a goroutine every `timeout` seconds
func (self *containerExporter) RecordMetadata() {
	go func() {
		for {
			containers := self.getContainers()
			hostname := self.getHostname()

			self.Metadata.Reset()
			for _, cnt := range containers {
				labels := utils.BuildLabels(cnt)
				self.Metadata.With(prometheus.Labels{
					"id":       labels.Id,
					"image":    labels.Image,
					"name":     labels.Name,
					"status":   labels.Status,
					"state":    labels.State,
					"nodename": hostname,
				})
			}

			time.Sleep(time.Duration(self.Timeout) * time.Second)
		}
	}()
}

// Initialize the containerExporter's Count GaugeVecs
func (self *containerExporter) initCountGauges() {
	self.Counts.Total = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "migratorius_docker_cnt_counts",
		Help: "Number of Docker containers detected on the node",
	},
		[]string{"nodename"},
	)
	self.Counts.Created = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "migratorius_docker_cnt_created",
		Help: "Number of Docker containers with status 'created'",
	},
		[]string{"nodename"},
	)
	self.Counts.Running = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "migratorius_docker_cnt_running",
		Help: "Number of Docker containers with status 'running'",
	},
		[]string{"nodename"},
	)
	self.Counts.Exited = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "migratorius_docker_cnt_exited",
		Help: "Number of Docker containers with status 'exited'",
	},
		[]string{"nodename"},
	)
}

// Initialize the containerExporter's Metadata GaugeVec
func (self *containerExporter) initMetadataGauge() {
	self.Metadata = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "migratorius_docker_cnt_metadata",
		Help: "Container metadata",
	},
		[]string{"id", "image", "name", "status", "state", "nodename"},
	)
}

// Get a list of existing containers via the Docker SDK
func (self *containerExporter) getContainers() []types.Container {
	containers, err := self.Client.ContainerList(self.Context, types.ContainerListOptions{All: true})
	if err != nil {
		log.Panic("Failed to get a container list")
	}
	return containers
}

// Get the hostname of the current node
func (self *containerExporter) getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal("Failed to get the hostname from the OS")
	}
	return hostname
}
