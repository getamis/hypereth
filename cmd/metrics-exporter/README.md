# metrics-exporter

A tool that exports metrics from an Ethereum client. It queries metrics from either HTTP, Websocket or IPC for every `period` and then passes the metrics to the backends.
Please note that you


## Supported backends

* [Prometheus](https://prometheus.io/)

  Expose all metrics through Prometheus gauge at **http://`host`:`port`/metrics**.

## Supported Ethereum client

* [go-ethereum](https://github.com/ethereum/go-ethereum)

  > Make sure the client is running with `--metrics` and the [debug](https://github.com/ethereum/go-ethereum/wiki/Management-APIs#debug), [admin](https://github.com/ethereum/go-ethereum/wiki/Management-APIs#admin), [txpool](https://github.com/ethereum/go-ethereum/wiki/Management-APIs#txpool) API is supported.

## Installing

See [README](../../../README.md)

## Build

See [README](../../../README.md)

## Usage

```
$ metrics-exporter --help
The Ethereum metrics exporter for Prometheus.

Usage:
  metrics-exporter [flags]

Flags:
      --eth.endpoint string     The Ethereum endpoint to connect to (default ":8546")
  -h, --help                    help for metrics-exporter
      --host string             The HTTP server listening address (default "localhost")
      --labels stringToString   The labels of metrics. For example: k1=v1,k2=v2 (default [])
      --namespace string        The namespace of metrics
      --period duration         The metrics update period (default 5s)
      --port int                The HTTP server listening port (default 9092)
```
