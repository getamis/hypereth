# metrics-exporter

A tool that exports metrics from an Ethereum client. It queries metrics from either HTTP, Websocket or IPC for every `period` and then passes the metrics to the backends.
Please note that you


## Supported backends

* [Prometheus](https://prometheus.io/)

  Expose all metrics through Prometheus gauge at **http://`host`:`port`/metrics**.

## Supported Ethereum client

* [go-ethereum](https://github.com/ethereum/go-ethereum)

  > Make sure the client is running with `--metrics` and the [debug](https://github.com/ethereum/go-ethereum/wiki/Management-APIs#debug) API is supported.

## Installing

See [README](../../../README.md)

## Build

See [README](../../../README.md)

## Usage

```
$ hypereth metrics-exporter --help
The Ethereum metrics exporter for Prometheus.

Usage:
  hypereth metrics-exporter [flags]

Flags:
      --eth.endpoint string   The Ethereum endpoint to connect to (default ":8546")
  -h, --help                  help for metrics-exporter
      --host string           The HTTP server listening address (default "localhost")
      --period duration       The metrics update period (default 5s)
      --port int              The HTTP server listening port (default 9092)
      --prefix string         The metrics name prefix

Global Flags:
      --config string   config file (default is .hypereth.yaml)
```
