# cnexporter
Prometheus exporter for certain Docker container metadata metrics

[![CodeQL](https://github.com/corvus-migratorius/cnexporter/actions/workflows/github-code-scanning/codeql/badge.svg)](https://github.com/corvus-migratorius/cnexporter/actions/workflows/github-code-scanning/codeql)

[![build](https://github.com/corvus-migratorius/cnexporter/actions/workflows/go.yml/badge.svg)](https://github.com/corvus-migratorius/cnexporter/actions/workflows/go.yml)

`cnexporter` uses Go Docker SDK to poll Docker API for a handful of container data.


## What it does

The exporter publishes several custom metrics via its `/metrics` endpoint:

- `cnexporter_containers_total`: Gauge, number of Docker containers that currently exist on the reporting system
- `cnexporter_containers_running` Gauge, number of Docker containers with state: `running`
- `cnexporter_containers_created` Gauge, number of containers with state: `created`
- `cnexporter_containers_exited` Gauge, number of containers with state: `exited`

- `cnexporter_containers_metadata` Gauge, always reports 0. This metric contains the following labels:
- - `id`
- - `image`: name of the registry image
- - `name`: name of the container itself
- - `status`: e.g. `Up 10 days (healthy)` (taken directly from the API)
- - `state`: `running`, `exited`, etc.
  
Also, all metrics expose the `nodename` label that is set to the reporting system's hostname (from an `os.Hostname()` call). This is done to allow matching with `node_uname_info` metric provided by https://github.com/prometheus/node_exporter and for other people like me who need this way of querying Prometheus.


## Rationale

I was frustrated with https://github.com/google/cadvisor, how convoluted and bloated it is. Also, sadly, https://github.com/prometheus-net/docker_exporter (which provided certain useful container metadata) was decomissioned. So I decided to write a dead-simple and lightweight tool that would give me the specific overview metrics that I need in my work.


## Installation

Grab a binary from the Releases page (Linux x86_64/amd64 only).

Alternatively, compile the exporter yourself (you will need at least Go 1.18).

```bash
make install
```

This will put `cnexporter` under `/usr/bin` by default (customizable via the `PREFIX` environment variable).

Or manually call `go build main.go`, compiling for whatever platform you want.


## Deployment

Just run it in whatever persistent way you prefer. If you want to go the `systemd` way, there's an example `.service` file in the repo.


## Arguments

As of now, there are but two options that you can set via CLI (aside from `--version`):

`--port`: set a port for publishing the Prometheus exporter metrics (defaul: `9200`);

`--timeout`: how often the tool will poll Docker API for information, in seconds (default: `15`).