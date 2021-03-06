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

	"github.com/ethereum/go-ethereum/common/hexutil"
)

// TxPoolStatus returns the hex encoding status of txpool.
// {
//   "pending": "0x1d5",
//   "queued": "0x1b7"
// }
func (ec *Client) TxPoolStatus(ctx context.Context) (map[string]hexutil.Uint, error) {
	r := make(map[string]hexutil.Uint)
	err := ec.c.CallContext(ctx, &r, "txpool_status")
	if err != nil {
		return nil, err
	}
	return r, nil
}
