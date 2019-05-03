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

package multiclient

import (
	"time"

	"github.com/getamis/sirius/log"
)

// Option represents a Client option
type Option func(*Client) error

// EthURLs represents static ethclient endpoints.
func EthURLs(urls []string) Option {
	return func(mc *Client) error {
		log.Info("EthClients from static list", "urls", urls)
		for _, url := range urls {
			mc.ClientMap().Add(url, nil)
		}
		return nil
	}
}

var (
	defaultRetryTimeout = 5 * time.Second
	defaultRetryDelay   = 1 * time.Second
)

type RetryConfig struct {
	// Limit is the total retry times. Set to 0 means retry time is according to the number of eth clients.
	Limit int
	// Timeout is the timeout for each retry. Set to 0 means use default timeout 5 seconds.
	Timeout time.Duration
	// Delay is the delay duration for each retry. Set to 0 means use default delay 1 second.
	Delay time.Duration
}

// WithRetryConfig configures the parameters for request retry.
func WithRetryConfig(retry RetryConfig) Option {
	return func(mc *Client) error {
		log.Info("Use given retry config", "retryLimit", retry.Limit, "retryTimeout", retry.Timeout, "retryDelay", retry.Delay)
		if retry.Timeout == 0 {
			log.Info("Use default retry timeout: 5 second")
			retry.Timeout = defaultRetryTimeout
		}
		if retry.Delay == 0 {
			log.Info("Use default retry delay: 1 second")
			retry.Delay = defaultRetryDelay
		}

		mc.requestRetryFunc = NewRetry(retry.Limit, retry.Timeout, retry.Delay)
		return nil
	}
}
