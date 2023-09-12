package main

import (
	"context"
	"fmt"
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
	cntStats = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "test_stat",
		Help: "This is a test Gauge metric",
	})
)

func recordMetrics() {
	go func() {
		cntStats.Set(0)
		for {
			time.Sleep(15 * time.Second)
		}
	}()
}

func main() {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		fmt.Printf(
			"%s : %s : %s : %s : %s\n",
			container.ID[:10],
			container.Image,
			container.Names[0],
			container.Status,
			container.State,
		)
	}

	recordMetrics()  // a coroutine

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":2112", nil))
}
