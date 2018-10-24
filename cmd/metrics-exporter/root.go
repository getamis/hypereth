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
	host        string
	port        int
	ethEndpoint string
	period      time.Duration
	namespace   string
	labels      map[string]string
)

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

// RootCmd represents the Prometheus metrics exporter
var RootCmd = &cobra.Command{
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
		metricsRegistry.SetNamespace(namespace)
		metricsRegistry.AppendLabels(labels)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go gaugeMetricsCollector(ctx, metricsRegistry.NewGauge,
			getETHMetrics(ctx, client, period),
			getBlockNumber(ctx, client, period),
			getPeerCount(ctx, client, period),
			getTxPoolStatus(ctx, client, period))

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
	RootCmd.Flags().StringVar(&host, "host", "localhost", "The HTTP server listening address")
	RootCmd.Flags().IntVar(&port, "port", 9092, "The HTTP server listening port")
	RootCmd.Flags().StringVar(&ethEndpoint, "eth.url", "ws://127.0.0.1:8546", "The Ethereum endpoint to connect to")
	RootCmd.Flags().DurationVar(&period, "period", 5*time.Second, "The metrics update period")
	RootCmd.Flags().StringVar(&namespace, "namespace", "", "The namespace of metrics")
	RootCmd.Flags().StringToStringVar(&labels, "labels", map[string]string{}, "The labels of metrics. For example: k1=v1,k2=v2")
}
