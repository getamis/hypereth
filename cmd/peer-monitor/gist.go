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
	"net/http"
	"regexp"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/getamis/sirius/log"
)

const (
	gistURL = "https://gist.githubusercontent.com/rfikki/a2ccdc1a31ff24884106da7b9e6a7453/raw/mainnet-peers-latest.txt"
)

var (
	enodeRegExp = regexp.MustCompile(`enode:\/\/([0-9]|[a-z]|[A-Z])+@[0-9]+(\.[0-9]+){3}:[0-9]+`)
)

func fetchFromGist(filter map[string]bool, max int) []*enode.Node {
	log.Trace("Start to fetch enodes from gist")
	resp, err := http.Get(gistURL)
	if err != nil {
		log.Error("Failed fetch node data from gist", "url", gistURL, "err", err)
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
	if len(enodes) >= max {
		enodes = enodes[:max]
	}
	log.Trace("Finished to fetch enodes from gist", "nodeCount", len(enodes))
	return enodes
}
