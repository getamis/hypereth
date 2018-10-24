# peer-monitor

A tool that monitor the peer status for an Ethereum client. It queries peer set from either HTTP, Websocket or IPC for every `period`. If the peer count is lower than requirement, add more peers to Ethereum client automatically.
Please note that you


## Sources of mainnet peers

* [Whitelist maintained by rfikki](https://gist.github.com/rfikki/a2ccdc1a31ff24884106da7b9e6a7453)
* [Ethernodes](https://www.ethernodes.org/network/1)

## Supported Ethereum client

* [go-ethereum](https://github.com/ethereum/go-ethereum)

  > Make sure the client supports the [admin](https://github.com/ethereum/go-ethereum/wiki/Management-APIs#admin) API.

## Installing

See [README](../../../README.md)

## Build

```
$ make peer-monitor
```

## Usage

```
$ peer-monitor --help
The Ethereum peer monitor. Need to open admin api for peer monitor

Usage:
  peer-monitor [flags]
  peer-monitor [command]

Available Commands:
  help        Help about any command
  once        once runs peer monitor once

Flags:
      --eth.url string              The Ethereum endpoint to connect to (default "ws://127.0.0.1:8546")
  -h, --help                        help for peer-monitor
      --monitor.duration duration   Monitor duration for eth peer set (default 1h0m0s)
      --peercount.max int           Maximum number of peer count (default 15)
      --peercount.min int           Minimum number of peer count (default 5)
```
