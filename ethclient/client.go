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

package ethclient

import (
	"context"
	"strings"

	"github.com/ethereum/go-ethereum/rpc"
)

// client defines typed wrappers for the Ethereum RPC API.
type client struct {
	c *rpc.Client
}

// Dial connects a client to the given URL.
func Dial(endpoint string) (*client, error) {
	return DialContext(context.Background(), endpoint)
}

// DialContext connects a client to the given URL with given context.
func DialContext(ctx context.Context, endpoint string) (*client, error) {
	if strings.HasPrefix(endpoint, "rpc:") || strings.HasPrefix(endpoint, "ipc:") {
		// Backwards compatibility with geth < 1.5 which required
		// these prefixes.
		endpoint = endpoint[4:]
	}

	c, err := rpc.DialContext(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	return NewClient(c), nil
}

// NewClient creates a client that uses the given RPC client.
func NewClient(c *rpc.Client) *client {
	return &client{c}
}

// Close closes an existing RPC connection.
func (ec *client) Close() {
	ec.c.Close()
}
