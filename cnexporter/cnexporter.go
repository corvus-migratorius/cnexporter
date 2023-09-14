package cnexporter

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
			cnt_created := float64(utils.CountByStatus(containers, "created"))
			cnt_running := float64(utils.CountByStatus(containers, "running"))
			cnt_exited := float64(utils.CountByStatus(containers, "exited"))

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
