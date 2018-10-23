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
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

// NetworkID returns the network ID (also known as the chain ID) for this chain.
func (ec *Client) NetworkID(ctx context.Context) (*big.Int, error) {
	version := new(big.Int)
	var ver string
	if err := ec.c.CallContext(ctx, &ver, "net_version"); err != nil {
		return nil, err
	}
	if _, ok := version.SetString(ver, 10); !ok {
		return nil, fmt.Errorf("invalid net_version result %q", ver)
	}
	return version, nil
}

// PeerCount returns the number of peers.
func (ec *Client) PeerCount(ctx context.Context) (uint64, error) {
	var result hexutil.Uint64
	if err := ec.c.CallContext(ctx, &result, "net_peerCount"); err != nil {
		return uint64(0), err
	}
	return uint64(result), nil
}
