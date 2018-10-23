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
	"strconv"
	"time"

	"github.com/getamis/sirius/metrics"
)

const (
	metricsNameSeperator = "_"
)

type GaugeCreater func(key string, opts ...metrics.Option) metrics.Gauge

type MetricsGetter func(results map[string]string)

func gaugeMetricsCollector(ctx context.Context, createGauge GaugeCreater, metricGetters ...MetricsGetter) error {
	gauges := make(map[string]metrics.Gauge)

	ticker := time.NewTicker(period)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			results := make(map[string]string)
			for _, getter := range metricGetters {
				getter(results)
			}

			for key, value := range results {
				var m metrics.Gauge

				if _, ok := gauges[key]; !ok {
					m = createGauge(key)
					gauges[key] = m
				} else {
					m = gauges[key]
				}

				if v, err := strconv.ParseFloat(value, 64); err == nil {
					m.Set(v)
				}
			}
		}
	}
}
