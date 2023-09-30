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

	"cnexporter/exporter"
)

func parseArgs(APPNAME string, VERSION string) (*int, *int) {
	parser := argparse.NewParser(APPNAME, "Publishes container metadata as a Prometheus exporter")
	port := parser.Int("p", "port", &argparse.Options{
		Help:    "Port for publishing the Prometheus exporter metrics",
		Default: 9200,
	})
	timeout := parser.Int("t", "timeout", &argparse.Options{
		Help:    "Timeout for polling Docker API (seconds)",
		Default: 15,
	})
	version := parser.Flag("V", "version", &argparse.Options{
		Help: "Display version number and quit",
	})

	// Parse input and display usage on error (equivalent to `--help`)
	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
	}

	if *version {
		fmt.Printf("%s: version %s\n", APPNAME, VERSION)
		os.Exit(0)
	}

	return port, timeout
}

func main() {
	APPNAME := "cnexporter"
	VERSION := "0.1.2"

	port, timeout := parseArgs(APPNAME, VERSION)

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
