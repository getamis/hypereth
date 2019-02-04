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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/getamis/sirius/log"
)

const (
	ClientGeth   = "Geth"
	ClientParity = "Parity"
)

const (
	fetchCount = 25
	fetchRound = 10
	clientType = ""

	drawField   = "{.DRAW}"
	lengthField = "{.LENGTH}"
	clientField = "{.CLIENT}"
	fetchURL    = "https://www.ethernodes.org/network/1/data?draw={.DRAW}&columns%5B0%5D%5Bdata%5D=id&columns%5B0%5D%5Bname%5D=&columns%5B0%5D%5Bsearchable%5D=true&columns%5B0%5D%5Borderable%5D=true&columns%5B0%5D%5Bsearch%5D%5Bvalue%5D=&columns%5B0%5D%5Bsearch%5D%5Bregex%5D=false&columns%5B1%5D%5Bdata%5D=host&columns%5B1%5D%5Bname%5D=&columns%5B1%5D%5Bsearchable%5D=true&columns%5B1%5D%5Borderable%5D=true&columns%5B1%5D%5Bsearch%5D%5Bvalue%5D=&columns%5B1%5D%5Bsearch%5D%5Bregex%5D=false&columns%5B2%5D%5Bdata%5D=port&columns%5B2%5D%5Bname%5D=&columns%5B2%5D%5Bsearchable%5D=true&columns%5B2%5D%5Borderable%5D=true&columns%5B2%5D%5Bsearch%5D%5Bvalue%5D=&columns%5B2%5D%5Bsearch%5D%5Bregex%5D=false&columns%5B3%5D%5Bdata%5D=country&columns%5B3%5D%5Bname%5D=&columns%5B3%5D%5Bsearchable%5D=true&columns%5B3%5D%5Borderable%5D=true&columns%5B3%5D%5Bsearch%5D%5Bvalue%5D=&columns%5B3%5D%5Bsearch%5D%5Bregex%5D=false&columns%5B4%5D%5Bdata%5D=clientId&columns%5B4%5D%5Bname%5D=&columns%5B4%5D%5Bsearchable%5D=true&columns%5B4%5D%5Borderable%5D=true&columns%5B4%5D%5Bsearch%5D%5Bvalue%5D=&columns%5B4%5D%5Bsearch%5D%5Bregex%5D=false&columns%5B5%5D%5Bdata%5D=client&columns%5B5%5D%5Bname%5D=&columns%5B5%5D%5Bsearchable%5D=true&columns%5B5%5D%5Borderable%5D=true&columns%5B5%5D%5Bsearch%5D%5Bvalue%5D=&columns%5B5%5D%5Bsearch%5D%5Bregex%5D=false&columns%5B6%5D%5Bdata%5D=clientVersion&columns%5B6%5D%5Bname%5D=&columns%5B6%5D%5Bsearchable%5D=true&columns%5B6%5D%5Borderable%5D=true&columns%5B6%5D%5Bsearch%5D%5Bvalue%5D=&columns%5B6%5D%5Bsearch%5D%5Bregex%5D=false&columns%5B7%5D%5Bdata%5D=os&columns%5B7%5D%5Bname%5D=&columns%5B7%5D%5Bsearchable%5D=true&columns%5B7%5D%5Borderable%5D=true&columns%5B7%5D%5Bsearch%5D%5Bvalue%5D=&columns%5B7%5D%5Bsearch%5D%5Bregex%5D=false&columns%5B8%5D%5Bdata%5D=lastUpdate&columns%5B8%5D%5Bname%5D=&columns%5B8%5D%5Bsearchable%5D=true&columns%5B8%5D%5Borderable%5D=true&columns%5B8%5D%5Bsearch%5D%5Bvalue%5D=&columns%5B8%5D%5Bsearch%5D%5Bregex%5D=false&order%5B0%5D%5Bcolumn%5D=8&order%5B0%5D%5Bdir%5D=desc&start=0&length={.LENGTH}&search%5Bvalue%5D={.CLIENT}&search%5Bregex%5D=false"
)

type ethernodeData struct {
	Nodes []*ethernode `json:"data"`
}

type ethernode struct {
	ID         string    `json:"id"`
	Host       string    `json:"host"`
	Port       int       `json:"port"`
	Client     string    `json:"client"`
	LastUpdate time.Time `json:"lastUpdate"`
}

func (n *ethernode) URL() string {
	return fmt.Sprintf("enode://%s@%s:%d", n.ID, n.Host, n.Port)
}

func fetchFromEthNodes(filter map[string]bool, max int) []*enode.Node {
	log.Trace("Start to fetch enodes from ethernode")
	totalFilter := make(map[string]bool)
	for k, v := range filter {
		totalFilter[k] = v
	}
	enodes := make([]*enode.Node, 0)
	for i := 0; i < fetchRound; i++ {
		newNodes := fetchFromEthNodesByRound(totalFilter, i)
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
	log.Trace("Finished to fetch enodes from ethernode", "nodeCount", len(enodes))
	return enodes
}

func fetchFromEthNodesByRound(filter map[string]bool, round int) []*enode.Node {
	log.Trace("Fetch enodes from ethernode by round", "round", round)
	queryURL := strings.Replace(fetchURL, drawField, fmt.Sprintf("%d", round+1), -1)
	queryURL = strings.Replace(queryURL, lengthField, fmt.Sprintf("%d", fetchCount), -1)
	queryURL = strings.Replace(queryURL, clientField, clientType, -1)

	resp, err := http.Get(queryURL)
	if err != nil {
		log.Error("Failed fetch node data", "url", queryURL, "err", err)
		return nil
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("Failed to read response body", "err", err)
		return nil
	}
	var data ethernodeData
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Error("Failed to json unmarshal", "err", err)
		return nil
	}

	isValid := func(n *ethernode) bool {
		if filter[n.ID] {
			return false
		}

		if net.ParseIP(n.Host) == nil {
			return false
		}

		switch n.Client {
		case ClientGeth, ClientParity:
		default:
			return false
		}
		return true
	}

	nodeCh := make(chan *enode.Node, len(data.Nodes))
	for _, n := range data.Nodes {
		if !isValid(n) {
			nodeCh <- nil
			continue
		}

		go func(n *ethernode) {
			ports := []int{30303}
			if n.Port != 30303 {
				ports = append(ports, n.Port)
			}

			for _, p := range ports {
				n.Port = p
				v4node, err := enode.ParseV4(n.URL())
				if err != nil {
					nodeCh <- nil
					return
				}
				err = dialNode(v4node)
				if err == nil {
					nodeCh <- v4node
					return
				}
			}
			nodeCh <- nil
		}(n)
	}
	enodes := make([]*enode.Node, 0)
	for i := 0; i < len(data.Nodes); i++ {
		n := <-nodeCh
		if n != nil {
			enodes = append(enodes, n)
		}
	}
	return enodes
}
