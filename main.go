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

	"docker-exporter/exporter"
)

func parseArgs() (*int, *int) {
	parser := argparse.NewParser("cnexporter", "Publishes container metadata as a Prometheus exporter")
	port := parser.Int("p", "port", &argparse.Options{
		Help:    "Port for publishing the Prometheus exporter",
		Default: 9200,
	})
	timeout := parser.Int("t", "timeout", &argparse.Options{
		Help:    "Timeout for polling Docker API (seconds)",
		Default: 15,
	})

	// Parse input and display usage on error (equivalent to `--help`)
	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
	}

	return port, timeout
}

func main() {
	port, timeout := parseArgs()

	log.Printf("Using port %d to publish /metrics\n", *port)
	log.Printf("Setting Docker API polling timeout to %d seconds\n", *timeout)

	context := context.Background()
	dclient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer dclient.Close()

	cnexporter := exporter.ContainerExporter(context, dclient, *timeout)

	cnexporter.RecordCounts()
	cnexporter.RecordMetadata()

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
