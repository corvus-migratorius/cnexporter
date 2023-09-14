package cnexporter

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
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

type ContainerExporter struct {
	Context  context.Context
	Client   *client.Client
	Counts   CntCounts
	Metadata *prometheus.GaugeVec
}

func (self *ContainerExporter) Initialize() {
	self.initCountGauges()
	self.initMetadataGauge()
}

func (self *ContainerExporter) RecordCounts() {
	go func() {
		for {
			containers := self.getContainers()
			hostname := self.getHostname()

			cnt_total := float64(len(containers))
			cnt_created := float64(countByStatus(containers, "created"))
			cnt_running := float64(countByStatus(containers, "running"))
			cnt_exited := float64(countByStatus(containers, "exited"))

			self.Counts.Total.With(prometheus.Labels{"nodename": hostname}).Set(cnt_total)
			self.Counts.Created.With(prometheus.Labels{"nodename": hostname}).Set(cnt_created)
			self.Counts.Running.With(prometheus.Labels{"nodename": hostname}).Set(cnt_running)
			self.Counts.Exited.With(prometheus.Labels{"nodename": hostname}).Set(cnt_exited)

			time.Sleep(15 * time.Second)
		}
	}()
}

func (self *ContainerExporter) RecordMetadata() {
	go func() {
		for {
			containers := self.getContainers()
			hostname := self.getHostname()

			self.Metadata.Reset()
			for _, cnt := range containers {
				labels := buildLabels(cnt)
				self.Metadata.With(prometheus.Labels{
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

func (self *ContainerExporter) initCountGauges() {
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

func (self *ContainerExporter) initMetadataGauge() {
	self.Metadata = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "migratorius_docker_cnt_metadata",
		Help: "Container metadata",
	},
		[]string{"id", "image", "name", "status", "state", "nodename"},
	)
}

func (self *ContainerExporter) getContainers() []types.Container {
	containers, err := self.Client.ContainerList(self.Context, types.ContainerListOptions{All: true})
	if err != nil {
		log.Panic("Failed to get a container list")
	}
	return containers
}

func (self *ContainerExporter) getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal("Failed to read the hostname")
	}
	return hostname
}

func countByStatus(containers []types.Container, state string) int {
	count := 0
	for _, cnt := range containers {
		if cnt.State == state {
			count++
		}
	}
	return count
}

func buildLabels(container types.Container) CntLabels {
	return CntLabels{
		container.ID,
		container.Image,
		strings.Trim(container.Names[0], "/"),
		container.Status,
		container.State,
	}
}
