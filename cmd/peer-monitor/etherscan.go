// Copyright 2018 AMIS Technologies
// This file is part of the hypereth library.
//
// The hypereth library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The hypereth library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the hypereth library. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"bufio"
	"fmt"
	"net/http"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/getamis/sirius/log"
)

const (
	maxEtherscanPage = 10
	etherscanURL     = "https://etherscan.io/nodetracker/nodes?p=%d"
)

func fetchFromEtherscan(filter map[string]bool, max int) []*enode.Node {
	log.Trace("Start to fetch enodes from etherscan")
	totalFilter := make(map[string]bool)
	for k, v := range filter {
		totalFilter[k] = v
	}
	enodes := make([]*enode.Node, 0)
	for i := 1; i <= maxEtherscanPage; i++ {
		newNodes := fetchFromEtherscanByPage(totalFilter, i)
		enodes = append(enodes, newNodes...)
		if len(enodes) >= max {
			enodes = enodes[:max]
			break
		}
		// update totalFilter for next round
		for _, n := range newNodes {
			totalFilter[n.ID().String()] = true
		}
	}
	log.Trace("Finished to fetch enodes from etherscan", "nodeCount", len(enodes))
	return enodes
}

func fetchFromEtherscanByPage(filter map[string]bool, page int) []*enode.Node {
	log.Trace("Fetch enodes from etherscan by page", "page", page)
	queryURL := fmt.Sprintf(etherscanURL, page)

	resp, err := http.Get(queryURL)
	if err != nil {
		log.Error("Failed fetch node data", "url", queryURL, "err", err)
		return nil
	}
	defer resp.Body.Close()

	candidates := make([]*enode.Node, 0)
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		txt := scanner.Text()
		nodeURLs := enodeRegExp.FindAllString(txt, -1)
		for _, nodeURL := range nodeURLs {
			n, err := enode.ParseV4(nodeURL)
			if err != nil {
				log.Error("Failed to parse enode url", "url", nodeURL, "err", err)
				continue
			}
			if !filter[n.ID().String()] {
				candidates = append(candidates, n)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		log.Error("Failed to read response body", "err", err)
		return nil
	}

	nodeCh := make(chan *enode.Node, len(candidates))
	for _, n := range candidates {
		go func(n *enode.Node) {
			err = dialNode(n)
			if err != nil {
				nodeCh <- nil
				return
			}
			nodeCh <- n
		}(n)
	}

	enodes := make([]*enode.Node, 0)
	for i := 0; i < len(candidates); i++ {
		n := <-nodeCh
		if n != nil {
			enodes = append(enodes, n)
		}
	}
	return enodes
}
