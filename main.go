package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/akamensky/argparse"
	"github.com/docker/docker/client"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"docker-exporter/cnexporter"
)

func main() {
	parser := argparse.NewParser("cnexporter", "Publishes container metadata as a Prometheus exporter")
	port := parser.Int("p", "port", &argparse.Options{
		Help:    "Port for publishing the Prometheus exporter",
		Default: 9200,
	})
	// timeout := parser.Int("t", "timeout", &argparse.Options{Help: "Timeout for polling Docker API"})
	// Parse input and display usage on error (equivalent to `--help`)
	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
	}

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
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
