package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"docker-exporter/exporter"
)

var (
	cntStats = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "migratorius_docker_cnt_metadata",
		Help: "Container metadata",
	},
		[]string{"id", "image", "name", "status", "state", "nodename"},
	)
	cntCountTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "migratorius_docker_cnt_counts",
		Help: "Number of Docker containers detected on the node",
	},
		[]string{"nodename"},
	)
	cntCountCreated = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "migratorius_docker_cnt_created",
		Help: "Number of Docker containers with status 'created'",
	},
		[]string{"nodename"},
	)
	cntCountRunning = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "migratorius_docker_cnt_running",
		Help: "Number of Docker containers with status 'running'",
	},
		[]string{"nodename"},
	)
	cntCountExited = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "migratorius_docker_cnt_exited",
		Help: "Number of Docker containers with status 'exited'",
	},
		[]string{"nodename"},
	)
)

func recordMetrics(dclient *client.Client, dcontext context.Context) {
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
			cnt_created := float64(exporter.CountByStatus(containers, "created"))
			cnt_running := float64(exporter.CountByStatus(containers, "running"))
			cnt_exited := float64(exporter.CountByStatus(containers, "exited"))

			cntStats.Reset()
			for _, cnt := range containers {
				labels := exporter.BuildLabels(cnt)
				cntStats.With(prometheus.Labels{
					"id":       labels.Id,
					"image":    labels.Image,
					"name":     labels.Name,
					"status":   labels.Status,
					"state":    labels.State,
					"nodename": hostname,
				})
			}

			cntCountTotal.With(prometheus.Labels{"nodename": hostname}).Set(cnt_total)
			cntCountCreated.With(prometheus.Labels{"nodename": hostname}).Set(cnt_created)
			cntCountRunning.With(prometheus.Labels{"nodename": hostname}).Set(cnt_running)
			cntCountExited.With(prometheus.Labels{"nodename": hostname}).Set(cnt_exited)

			time.Sleep(15 * time.Second)
		}
	}()
}

func main() {
	dcontext := context.Background()
	dclient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer dclient.Close()

	recordMetrics(dclient, dcontext) // a coroutine

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":2112", nil))
}
