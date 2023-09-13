package main

import (
	"context"
	// "fmt"
	"log"
	"net/http"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// cntStats = promauto.NewGaugeVec(prometheus.GaugeOpts{
	// 	Name: "test_stat",
	// 	Help: "This is a test Gauge metric",
	// },
	// 	[]string{"id", "image", "name", "status", "state", "epoch"},
	// )
	cntCountTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "migratorius_exporter_containers_total",
		Help: "Number of Docker containers detected on the node",
	},
		[]string{"nodename"},
	)
	cntCountCreated = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "migratorius_exporter_containers_created",
		Help: "Number of Docker containers with status 'created'",
	},
		[]string{"nodename"},
	)
	cntCountRunning = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "migratorius_exporter_containers_running",
		Help: "Number of Docker containers with status 'running'",
	},
		[]string{"nodename"},
	)
	cntCountExited = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "migratorius_exporter_containers_exited",
		Help: "Number of Docker containers with status 'exited'",
	},
		[]string{"nodename"},
	)
)

func countByStatus(containers []types.Container, state string) int {
	count := 0
	for _, cnt := range containers {
		if cnt.State == state {
			count++
		}
	}
	return count
}

func recordMetrics(dclient *client.Client, dcontext context.Context) {
	go func() {
		for {
			containers, err := dclient.ContainerList(dcontext, types.ContainerListOptions{All: true})
			if err != nil {
				panic(err)
			}
			cnt_total := float64(len(containers))
			cnt_created := float64(countByStatus(containers, "created"))
			cnt_running := float64(countByStatus(containers, "running"))
			cnt_exited := float64(countByStatus(containers, "exited"))

			// cntStats.With(prometheus.Labels{
			// 	"id":     "123",
			// 	"image":  "test",
			// 	"name":   "Test",
			// 	"status": "Created",
			// 	"state":  "created",
			// 	"epoch":   fmt.Sprint(time.Now().Unix()),
			// })
			cntCountTotal.With(prometheus.Labels{"nodename": "test"}).Set(cnt_total)
			cntCountCreated.With(prometheus.Labels{"nodename": "test"}).Set(cnt_created)
			cntCountRunning.With(prometheus.Labels{"nodename": "test"}).Set(cnt_running)
			cntCountExited.With(prometheus.Labels{"nodename": "test"}).Set(cnt_exited)

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

	// for _, container := range containers {
	// 	fmt.Printf(
	// 		"%s : %s : %s : %s : %s\n",
	// 		container.ID[:10],
	// 		container.Image,
	// 		container.Names[0],
	// 		container.Status,
	// 		container.State,
	// 	)
	// }

	recordMetrics(dclient, dcontext) // a coroutine

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":2112", nil))
}
