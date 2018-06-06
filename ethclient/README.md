eth-client
==========

[![License: LGPL v3](https://img.shields.io/badge/License-LGPL%20v3-blue.svg)](https://www.gnu.org/licenses/lgpl-3.0)

A Golang client library to communicate with Ethereum RPC server.
* Implements most of JSON-RPC methods and several client-specific methods.
* Provides a high-level interface to **propose/get validators** on Istanbul blockchain.
* Provides a high-level interface to **create private contracts** on Quorum blockchain.

Usage
-----
```golang
package main

import (
	"context"
	"fmt"

	"github.com/getamis/eth-client/client"
)

func main() {
	url := "http://127.0.0.1:8545"
	client, err := client.Dial(url)
	if err != nil {
		fmt.Println("Failed to dial, url: ", url, ", err: ", err)
		return
	}

	err = client.StartMining(context.Background())
	if err != nil {
		fmt.Println("Failed to start mining, err: ", err)
		return
	}
	fmt.Println("start mining")
}

```

Implemented JSON-RPC methods
----------------------------

* admin_addPeer
* admin_adminPeers
* admin_nodeInfo
* eth_blockNumber
* eth_sendRawTransaction
* eth_getBlockByHash
* eth_getBlockByNumber
* eth_getBlockByHash
* eth_getBlockByNumber
* eth_getTransactionByHash
* eth_getBlockTransactionCountByHash
* eth_getTransactionByBlockHashAndIndex
* eth_getTransactionReceipt
* eth_syncing
* eth_getBalance
* eth_getStorageAt
* eth_getCode
* eth_getBlockTransactionCountByNumber
* eth_call
* eth_gasPrice
* eth_estimateGas
* eth_sendRawTransaction
* miner_startMining
* miner_stopMining
* net_version
* eth_getLogs
* debug_metrics

### Istanbul-only JSON-RPC methods
To use these methods, make sure that
* Server is running on [Istanbul consensus](https://github.com/ethereum/EIPs/issues/650).
* Connect to server through `istanbul.Dial` function (not the original [Geth client](https://github.com/ethereum/go-ethereum/tree/master/ethclient)).

Methods:

* istanbul_getValidators
* istanbul_propose

Reference
---------

* https://github.com/ethereum/go-ethereum
* https://github.com/ethereum/wiki/wiki/JSON-RPC
* https://github.com/ethereum/EIPs/issues/650
* https://github.com/jpmorganchase/quorum
* https://github.com/getamis/istanbul-tools
