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

package metrics

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/getamis/sirius/log"
	"github.com/getamis/sirius/metrics"
	"github.com/spf13/cobra"

	"github.com/getamis/hypereth/ethclient"
)

const (
	metricsPath = "/metrics"
)

var (
	host          string
	port          int
	ethEndpoint   string
	period        time.Duration
	metricsPrefix string
)

// PrometheusCommand represents the Prometheus metrics exporter
var PrometheusCommand = &cobra.Command{
	Use:   "metrics-exporter",
	Short: "The Ethereum metrics exporter for Prometheus",
	Long:  `The Ethereum metrics exporter for Prometheus.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := ethclient.Dial(ethEndpoint)
		if err != nil {
			log.Error("Failed to connect to Ethereum", "err", err)
			return err
		}
		defer client.Close()

		metricsRegistry := metrics.NewPrometheusRegistry()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go gaugeMetricsCollector(ctx, func(key string, opts ...metrics.Option) metrics.Gauge {
			return metricsRegistry.NewGauge(key, opts...)
		}, client.Metrics, metricsPrefix)

		mux := http.NewServeMux()
		mux.Handle(metricsPath, metricsRegistry)
		srv := &http.Server{
			Addr:    fmt.Sprintf("%s:%d", host, port),
			Handler: mux,
		}

		go func() {
			sigs := make(chan os.Signal, 1)
			signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
			defer signal.Stop(sigs)
			log.Debug("Shutting down", "signal", <-sigs)
			srv.Shutdown(context.Background())
		}()

		log.Info("Starting metrics exporter", "endpoint", fmt.Sprintf("http://%s:%d%s", host, port, metricsPath))

		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			return err
		}

		return nil
	},
}

func init() {
	PrometheusCommand.Flags().StringVar(&host, "host", "localhost", "The HTTP server listening address")
	PrometheusCommand.Flags().IntVar(&port, "port", 9092, "The HTTP server listening port")
	PrometheusCommand.Flags().StringVar(&ethEndpoint, "eth.endpoint", ":8546", "The Ethereum endpoint to connect to")
	PrometheusCommand.Flags().DurationVar(&period, "period", 5*time.Second, "The metrics update period")
	PrometheusCommand.Flags().StringVar(&metricsPrefix, "prefix", "", "The metrics name prefix")
}
