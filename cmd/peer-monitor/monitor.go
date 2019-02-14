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
	"context"
	"errors"
	"net"
	"time"

	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/getamis/hypereth/ethclient"
	"github.com/getamis/sirius/log"
)

const (
	ctxTimeout   = 10 * time.Second
	retryTimeout = 3 * time.Second
	dialTimeout  = 5 * time.Second
)

var (
	dialer = p2p.TCPDialer{Dialer: &net.Dialer{Timeout: dialTimeout}}
)

type fetchFn func(filter map[string]bool, max int) []*enode.Node

type PeerMonitor struct {
	ethURL       string
	minPeerCount int
	maxPeerCount int
	fetcher      []fetchFn
	quit         chan struct{}
}

func NewPeerMonitor(ethURL string, minPeerCount, maxPeerCount int) *PeerMonitor {
	return &PeerMonitor{
		ethURL:       ethURL,
		minPeerCount: minPeerCount,
		maxPeerCount: maxPeerCount,
		quit:         make(chan struct{}),
		fetcher: []fetchFn{
			fetchFromGist,
			fetchFromEtherscan,
		},
	}
}

func (m *PeerMonitor) Run(monitorDuration time.Duration) error {
	// schedule run and force run immediately
	timer := time.NewTimer(0)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			log.Info("Start to check peer set")
			duration := monitorDuration
			err := m.RunOnce()
			if err != nil {
				log.Error("Failed to check peer set, retry", "err", err)
				duration = retryTimeout
			}
			timer.Reset(duration)
		case <-m.quit:
			return nil
		}
	}
}

func (m *PeerMonitor) Stop() {
	close(m.quit)
}

func (m *PeerMonitor) RunOnce() error {
	dialCtx, dialCancel := context.WithTimeout(context.Background(), ctxTimeout)
	defer dialCancel()
	ethClient, err := ethclient.DialContext(dialCtx, m.ethURL)
	if err != nil {
		return err
	}
	defer ethClient.Close()

	peersCtx, peersCancel := context.WithTimeout(context.Background(), ctxTimeout)
	defer peersCancel()
	peers, err := ethClient.AdminPeers(peersCtx)
	if err != nil {
		return err
	}

	log.Info("Current peers", "count", len(peers))
	if len(peers) > m.minPeerCount {
		log.Info("No need to discover nodes", "minPeerCount", m.minPeerCount)
		return nil
	}

	nodes := m.fetchNodes(peers, m.maxPeerCount)
	if len(nodes) == 0 {
		log.Error("empty node list")
		return errors.New("empty node list")
	}

	log.Trace("Start to batch add peer", "nodeCount", len(nodes))
	addPeerCtx, addPeerCancel := context.WithTimeout(context.Background(), ctxTimeout)
	defer addPeerCancel()
	err = ethClient.BatchAddPeer(addPeerCtx, nodes)
	if err != nil {
		log.Error("Failed to batch add peer", "err", err)
	}

	log.Info("Finish add peer")
	return nil
}

func (m *PeerMonitor) fetchNodes(curPeers []*p2p.PeerInfo, maxPeerCount int) []string {
	exists := make(map[string]bool)
	for _, p := range curPeers {
		exists[p.ID] = true
	}

	dist := maxPeerCount - len(curPeers)
	enodes := []string{}

	addNodes := func(candidates []*enode.Node) {
		for _, c := range candidates {
			exists[c.ID().String()] = true
			enodes = append(enodes, c.String())
		}
	}

	for _, f := range m.fetcher {
		addNodes(f(exists, dist))
		dist = dist - len(enodes)
		if dist <= 0 {
			break
		}
	}
	return enodes
}

func dialNode(node *enode.Node) error {
	conn, err := dialer.Dial(node)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}
