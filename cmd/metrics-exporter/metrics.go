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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/getamis/sirius/log"
	"github.com/tidwall/gjson"

	"github.com/getamis/hypereth/ethclient"
)

const (
	blockNumberMetricName        = "block_number"
	peerCountMetricName          = "peer_count"
	txPoolStatusMetricNamePrefix = "txpool_status"
)

func getETHMetrics(ctx context.Context, c *ethclient.Client, timeout time.Duration) MetricsGetter {
	return func(results map[string]string) {
		getCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		m, err := c.Metrics(getCtx)
		if err != nil {
			log.Error("Failed to retrieve metrics", "err", err)
			return
		}
		parseJSONMetrics(m, "", results)
	}
}

func getBlockNumber(ctx context.Context, c *ethclient.Client, timeout time.Duration) MetricsGetter {
	return func(results map[string]string) {
		getCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		b, err := c.HeaderByNumber(getCtx, nil)
		if err != nil {
			log.Error("Failed to retrieve latest block header", "err", err)
			return
		}
		results[blockNumberMetricName] = b.Number.String()
	}
}

func getPeerCount(ctx context.Context, c *ethclient.Client, timeout time.Duration) MetricsGetter {
	return func(results map[string]string) {
		getCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		cnt, err := c.PeerCount(getCtx)
		if err != nil {
			log.Error("Failed to retrieve peer count", "err", err)
			return
		}
		results[peerCountMetricName] = fmt.Sprintf("%d", cnt)
	}
}

func getTxPoolStatus(ctx context.Context, c *ethclient.Client, timeout time.Duration) MetricsGetter {
	return func(results map[string]string) {
		getCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		m, err := c.TxPoolStatus(getCtx)
		if err != nil {
			log.Error("Failed to retrieve metrics", "err", err)
			return
		}
		parseJSONMetrics(m, txPoolStatusMetricNamePrefix, results)
	}
}

func parseJSONMetrics(raw map[string]interface{}, prefix string, results map[string]string) {
	rawJSON, _ := json.Marshal(raw)
	expandJSON(gjson.Parse(string(rawJSON)), prefix, results, func(ks ...string) string {
		var keys []string
		for _, k := range ks {
			if k != "" {
				keys = append(keys, strings.ToLower(k))
			}
		}

		return strings.Join(keys, metricsNameSeperator)
	})
}
