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
	"context"
	"time"

	"github.com/ethereum/go-ethereum/rpc"
)

// RetryFunc is the function we retry
type RetryFunc func(ctx context.Context, rpcClient *rpc.Client) (retry bool, err error)

// NewRetry is an util function to retry with retry limit, timeout and delay
func NewRetry(retryLimit int, retryTimeout, retryDelay time.Duration) func(context.Context, []*rpc.Client, RetryFunc) error {
	return func(ctx context.Context, rpcClients []*rpc.Client, fn RetryFunc) error {
		limit := retryLimit
		if limit == 0 {
			limit = len(rpcClients)
		}
		return Retry(ctx, limit, retryTimeout, retryDelay, rpcClients, fn)
	}
}

// Retry retries the RetryFunc with limit and timeout and wait some delay during each retry
func Retry(ctx context.Context, retryLimit int, retryTimeout, retryDelay time.Duration, rpcClients []*rpc.Client, fn RetryFunc) error {
	timer := time.NewTimer(0)
	defer timer.Stop()

	attempt := 0
	for {
		idx := attempt % len(rpcClients)

		c, cancel := context.WithTimeout(ctx, retryTimeout)
		retry, err := fn(c, rpcClients[idx])
		cancel()

		if !retry || err == nil {
			return err
		}

		attempt++
		if attempt == retryLimit {
			return err
		}

		timer.Reset(retryDelay)
		select {
		case <-timer.C:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
