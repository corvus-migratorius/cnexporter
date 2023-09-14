package main

import (
	"context"
	"log"
	"net/http"

	"github.com/docker/docker/client"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"docker-exporter/cnexporter"
)

func main() {
	context := context.Background()
	dclient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer dclient.Close()

	exporter := cnexporter.ContainerExporter(context, dclient)

	exporter.RecordCounts()
	exporter.RecordMetadata()

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":2112", nil))
}
