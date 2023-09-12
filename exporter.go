package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

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

	http.Handle("/metrics", promhttp.Handler())
    http.ListenAndServe(":2112", nil)
}
