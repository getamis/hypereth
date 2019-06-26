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
	"errors"
	"math/big"
	"sync"
	"time"

	"github.com/cskr/pubsub"
	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/getamis/sirius/log"
)

const (
	dialTimeout = 5 * time.Second
	retryPeriod = 10 * time.Second

	// newAvailableClientTopic represents an topic name for an eth-client is created.
	newAvailableClientTopic = "newAvailableClient"
	// pubSubCapacity represents the channel size to received pubSub event.
	pubSubCapacity = 10
)

var (
	ErrInvalidTypeCast = errors.New("invalid type cast")
	ErrNoEthClient     = errors.New("no eth client")
)

type Client struct {
	ctx              context.Context
	cancel           context.CancelFunc
	rpcClientMap     *Map
	newClientCh      chan string
	pubSub           *pubsub.PubSub
	retrydialWg      sync.WaitGroup
	requestRetryFunc func(context.Context, []*rpc.Client, RetryFunc) error
}

func New(ctx context.Context, opts ...Option) (*Client, error) {
	newClientCh := make(chan string)
	// create client own context to control the internal go routines
	myCtx, myCancel := context.WithCancel(context.Background())
	mc := &Client{
		ctx:              myCtx,
		cancel:           myCancel,
		rpcClientMap:     NewMap(newClientCh),
		newClientCh:      newClientCh,
		pubSub:           pubsub.New(pubSubCapacity),
		requestRetryFunc: NewRetry(0, defaultRetryTimeout, defaultRetryDelay),
	}

	var newErr error
	defer func() {
		if newErr != nil {
			// cancel go routines
			mc.Close()
		}
	}()

	for _, opt := range opts {
		if newErr = opt(mc); newErr != nil {
			return nil, newErr
		}
	}

	// Dial each eth client
	mc.DialClients(ctx)

	// Return error if have no available eth client.
	if len(mc.rpcClientMap.List()) == 0 {
		log.Warn("There is no available eth client while creating multiclient")
	}

	mc.retrydialWg.Add(1)
	go mc.retrydial()

	return mc, nil
}

// Close closes an existing RPC connection.
func (mc *Client) Close() {
	// stop go routines
	mc.cancel()
	mc.retrydialWg.Wait()
	mc.pubSub.Shutdown()
	clients := mc.rpcClientMap.List()
	for _, c := range clients {
		c.Close()
	}
}

func (mc *Client) Context() context.Context {
	return mc.ctx
}

func (mc *Client) ClientMap() *Map {
	return mc.rpcClientMap
}

func (mc *Client) EthClients() []*ethclient.Client {
	clients := mc.rpcClientMap.List()
	ethClients := make([]*ethclient.Client, len(clients))
	for i, c := range clients {
		ethClients[i] = ethclient.NewClient(c)
	}
	return ethClients
}

func (mc *Client) RPCClients() []*rpc.Client {
	return mc.rpcClientMap.List()
}

// Blockchain Access

// BlockByHash returns the given full block.
//
// Note that loading full blocks requires two requests. Use HeaderByHash
// if you don't need all transactions or uncle headers.
func (mc *Client) BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error) {
	clients := mc.rpcClientMap.List()
	if len(clients) == 0 {
		return nil, ErrNoEthClient
	}

	var result *types.Block
	var errs []error

	finalErr := mc.requestRetryFunc(ctx, clients, func(ctx context.Context, rpcClient *rpc.Client) (bool, error) {
		ec := ethclient.NewClient(rpcClient)
		var err error
		result, err = ec.BlockByHash(ctx, hash)
		if err != nil {
			errs = append(errs, err)
		}
		return true, err
	})
	if finalErr != nil {
		log.Debug("Failed to get block by hash", "hash", hash.Hex(), "finalErr", finalErr, "errs", errs)
		return nil, finalErr
	}
	return result, nil
}

// BlockByNumber returns a block from the current canonical chain. If number is nil, the
// latest known block is returned.
//
// Note that loading full blocks requires two requests. Use HeaderByNumber
// if you don't need all transactions or uncle headers.
func (mc *Client) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	clients := mc.rpcClientMap.List()
	if len(clients) == 0 {
		return nil, ErrNoEthClient
	}

	var result *types.Block
	var errs []error

	finalErr := mc.requestRetryFunc(ctx, clients, func(ctx context.Context, rpcClient *rpc.Client) (bool, error) {
		ec := ethclient.NewClient(rpcClient)
		var err error
		result, err = ec.BlockByNumber(ctx, number)
		if err != nil {
			errs = append(errs, err)
		}
		return true, err
	})
	if finalErr != nil {
		log.Debug("Failed to get block by number", "number", number.String(), "finalErr", finalErr, "errs", errs)
		return nil, finalErr
	}
	return result, nil
}

// HeaderByHash returns the block header with the given hash.
func (mc *Client) HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error) {
	clients := mc.rpcClientMap.List()
	if len(clients) == 0 {
		return nil, ErrNoEthClient
	}

	var result *types.Header
	var errs []error

	finalErr := mc.requestRetryFunc(ctx, clients, func(ctx context.Context, rpcClient *rpc.Client) (bool, error) {
		ec := ethclient.NewClient(rpcClient)
		var err error
		result, err = ec.HeaderByHash(ctx, hash)
		if err != nil {
			errs = append(errs, err)
		}
		return true, err
	})
	if finalErr != nil {
		log.Debug("Failed to get block header by hash", "hash", hash.Hex(), "finalErr", finalErr, "errs", errs)
		return nil, finalErr
	}
	return result, nil
}

// HeaderByNumber returns a block header from the current canonical chain. If number is
// nil, the latest known header is returned.
func (mc *Client) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	clients := mc.rpcClientMap.List()
	if len(clients) == 0 {
		return nil, ErrNoEthClient
	}

	var result *types.Header
	var errs []error

	finalErr := mc.requestRetryFunc(ctx, clients, func(ctx context.Context, rpcClient *rpc.Client) (bool, error) {
		ec := ethclient.NewClient(rpcClient)
		var err error
		result, err = ec.HeaderByNumber(ctx, number)
		if err != nil {
			errs = append(errs, err)
		}
		return true, err
	})
	if finalErr != nil {
		log.Debug("Failed to get block header by number", "number", number.String(), "finalErr", finalErr, "errs", errs)
		return nil, finalErr
	}
	return result, nil
}

// TransactionByHash returns the transaction with the given hash.
func (mc *Client) TransactionByHash(ctx context.Context, hash common.Hash) (*types.Transaction, bool, error) {
	clients := mc.rpcClientMap.List()
	if len(clients) == 0 {
		return nil, false, ErrNoEthClient
	}

	var result *types.Transaction
	var isPending bool
	var errs []error

	finalErr := mc.requestRetryFunc(ctx, clients, func(ctx context.Context, rpcClient *rpc.Client) (bool, error) {
		ec := ethclient.NewClient(rpcClient)
		var err error
		result, isPending, err = ec.TransactionByHash(ctx, hash)
		if err != nil {
			errs = append(errs, err)
		}
		return true, err
	})
	if finalErr != nil {
		log.Debug("Failed to get transaction by hash", "hash", hash.Hex(), "finalErr", finalErr, "errs", errs)
		return nil, false, finalErr
	}
	return result, isPending, nil
}

// State Access

// BalanceAt returns the wei balance of the given account.
// The block number can be nil, in which case the balance is taken from the latest known block.
func (mc *Client) BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
	clients := mc.rpcClientMap.List()
	if len(clients) == 0 {
		return nil, ErrNoEthClient
	}

	var result *big.Int
	var errs []error

	finalErr := mc.requestRetryFunc(ctx, clients, func(ctx context.Context, rpcClient *rpc.Client) (bool, error) {
		ec := ethclient.NewClient(rpcClient)
		var err error
		result, err = ec.BalanceAt(ctx, account, blockNumber)
		if err != nil {
			errs = append(errs, err)
		}
		return true, err
	})
	if finalErr != nil {
		log.Debug("Failed to get balance", "account", account.Hex(), "blockNumber", blockNumber.String(), "finalErr", finalErr, "errs", errs)
		return nil, finalErr
	}
	return result, nil
}

// CodeAt returns the contract code of the given account.
// The block number can be nil, in which case the code is taken from the latest known block.
func (mc *Client) CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error) {
	clients := mc.rpcClientMap.List()
	if len(clients) == 0 {
		return nil, ErrNoEthClient
	}

	var result []byte
	var errs []error

	finalErr := mc.requestRetryFunc(ctx, clients, func(ctx context.Context, rpcClient *rpc.Client) (bool, error) {
		ec := ethclient.NewClient(rpcClient)
		var err error
		result, err = ec.CodeAt(ctx, account, blockNumber)
		if err != nil {
			errs = append(errs, err)
		}
		return true, err
	})
	if finalErr != nil {
		log.Debug("Failed to get code", "account", account.Hex(), "blockNumber", blockNumber.String(), "finalErr", finalErr, "errs", errs)
		return nil, finalErr
	}
	return result, nil
}

// NonceAt returns the account nonce of the given account.
// The block number can be nil, in which case the nonce is taken from the latest known block.
func (mc *Client) NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error) {
	clients := mc.rpcClientMap.List()
	if len(clients) == 0 {
		return 0, ErrNoEthClient
	}

	var result uint64
	var errs []error

	finalErr := mc.requestRetryFunc(ctx, clients, func(ctx context.Context, rpcClient *rpc.Client) (bool, error) {
		ec := ethclient.NewClient(rpcClient)
		var err error
		result, err = ec.NonceAt(ctx, account, blockNumber)
		if err != nil {
			errs = append(errs, err)
		}
		return true, err
	})
	if finalErr != nil {
		log.Debug("Failed to get nonce", "account", account.Hex(), "blockNumber", blockNumber.String(), "finalErr", finalErr, "errs", errs)
		return uint64(0), finalErr
	}
	return result, nil
}

// Pending State

// PendingBalanceAt returns the wei balance of the given account in the pending state.
func (mc *Client) PendingBalanceAt(ctx context.Context, account common.Address) (*big.Int, error) {
	clients := mc.rpcClientMap.List()
	if len(clients) == 0 {
		return nil, ErrNoEthClient
	}

	var result *big.Int
	var errs []error

	finalErr := mc.requestRetryFunc(ctx, clients, func(ctx context.Context, rpcClient *rpc.Client) (bool, error) {
		ec := ethclient.NewClient(rpcClient)
		var err error
		result, err = ec.PendingBalanceAt(ctx, account)
		if err != nil {
			errs = append(errs, err)
		}
		return true, err
	})
	if finalErr != nil {
		log.Debug("Failed to get pending balance", "account", account.Hex(), "finalErr", finalErr, "errs", errs)
		return nil, finalErr
	}
	return result, nil
}

// PendingNonceAt returns the account nonce of the given account in the pending state.
// This is the nonce that should be used for the next transaction.
func (mc *Client) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	clients := mc.rpcClientMap.List()
	if len(clients) == 0 {
		return 0, ErrNoEthClient
	}

	var result uint64
	var errs []error

	finalErr := mc.requestRetryFunc(ctx, clients, func(ctx context.Context, rpcClient *rpc.Client) (bool, error) {
		ec := ethclient.NewClient(rpcClient)
		var err error
		result, err = ec.PendingNonceAt(ctx, account)
		if err != nil {
			errs = append(errs, err)
		}
		return true, err
	})
	if finalErr != nil {
		log.Debug("Failed to get pending nonce", "account", account.Hex(), "finalErr", finalErr, "errs", errs)
		return uint64(0), finalErr
	}
	return result, nil
}

// Contract Calling

// CallContract executes a message call transaction, which is directly executed in the VM
// of the node, but never mined into the blockchain.
//
// blockNumber selects the block height at which the call runs. It can be nil, in which
// case the code is taken from the latest known block. Note that state from very old
// blocks might not be available.
func (mc *Client) CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	clients := mc.rpcClientMap.List()
	if len(clients) == 0 {
		return nil, ErrNoEthClient
	}

	var result []byte
	var errs []error

	finalErr := mc.requestRetryFunc(ctx, clients, func(ctx context.Context, rpcClient *rpc.Client) (bool, error) {
		ec := ethclient.NewClient(rpcClient)
		var err error
		result, err = ec.CallContract(ctx, msg, blockNumber)
		if err != nil {
			errs = append(errs, err)
		}
		return true, err
	})
	if finalErr != nil {
		log.Debug("Failed to call contract", "from", msg.From.Hex(), "to", msg.To.Hex(), "blockNumber", blockNumber.String(), "finalErr", finalErr, "errs", errs)
		return nil, finalErr
	}
	return result, nil
}

// PendingCallContract executes a message call transaction using the EVM.
// The state seen by the contract call is the pending state.
func (mc *Client) PendingCallContract(ctx context.Context, msg ethereum.CallMsg) ([]byte, error) {
	clients := mc.rpcClientMap.List()
	if len(clients) == 0 {
		return nil, ErrNoEthClient
	}

	var result []byte
	var errs []error

	finalErr := mc.requestRetryFunc(ctx, clients, func(ctx context.Context, rpcClient *rpc.Client) (bool, error) {
		ec := ethclient.NewClient(rpcClient)
		var err error
		result, err = ec.PendingCallContract(ctx, msg)
		if err != nil {
			errs = append(errs, err)
		}
		return true, err
	})
	if finalErr != nil {
		log.Debug("Failed to call contract with pending state", "from", msg.From.Hex(), "to", msg.To.Hex(), "finalErr", finalErr, "errs", errs)
		return nil, finalErr
	}
	return result, nil
}

// SendTransaction injects a signed transaction into the pending pool for execution.
// Return all errors if multiple errors have occurred.
//
// If the transaction was a contract creation use the TransactionReceipt method to get the
// contract address after the transaction has been mined.
func (mc *Client) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	clients := mc.rpcClientMap.Map()
	if len(clients) == 0 {
		return ErrNoEthClient
	}

	respCh := make(chan error, len(clients))

	for url, c := range clients {
		go func(url string, c *rpc.Client) {
			ec := ethclient.NewClient(c)
			err := ec.SendTransaction(ctx, tx)
			clientErr := NewClientError(url, err)
			respCh <- clientErr
		}(url, c)
	}

	var errs []error
	for i := 0; i < len(clients); i++ {
		respErr := <-respCh
		if respErr != nil {
			errs = append(errs, respErr)
		}
	}

	if len(errs) == len(clients) {
		log.Debug("Failed to send transaction", "txHash", tx.Hash().Hex(), "errs", errs)
		return NewMultipleError(errs)
	}

	return nil
}

// CallContext performs a JSON-RPC call with the given arguments. If the context is
// canceled before the call has successfully returned, CallContext returns immediately.
//
// The result must be a pointer so that package json can unmarshal into it. You
// can also pass nil, in which case the result is ignored.
func (mc *Client) CallContext(ctx context.Context, result interface{}, method string, args ...interface{}) error {
	clients := mc.rpcClientMap.List()
	if len(clients) == 0 {
		return ErrNoEthClient
	}

	var errs []error

	finalErr := mc.requestRetryFunc(ctx, clients, func(ctx context.Context, rpcClient *rpc.Client) (bool, error) {
		err := rpcClient.CallContext(ctx, result, method, args...)
		if err != nil {
			errs = append(errs, err)
		}
		return true, err
	})
	if finalErr != nil {
		log.Debug("Failed to perform a JSON-RPC call", "finalErr", finalErr, "errs", errs)
		return finalErr
	}
	return nil
}

// BatchCall sends all given requests as a single batch and waits for the server
// to return a response for all of them. The wait duration is bounded by the
// context's deadline.
//
// In contrast to CallContext, BatchCallContext only returns errors that have occurred
// while sending the request. Any error specific to a request is reported through the
// Error field of the corresponding BatchElem.
//
// Note that batch calls may not be executed atomically on the server side.
func (mc *Client) BatchCallContext(ctx context.Context, b []rpc.BatchElem) error {
	clients := mc.rpcClientMap.List()
	if len(clients) == 0 {
		return ErrNoEthClient
	}

	var errs []error

	finalErr := mc.requestRetryFunc(ctx, clients, func(ctx context.Context, rpcClient *rpc.Client) (bool, error) {
		err := rpcClient.BatchCallContext(ctx, b)
		if err != nil {
			errs = append(errs, err)
		}
		return true, err
	})
	if finalErr != nil {
		log.Debug("Failed to perform batch JSON-RPC calls", "finalErr", finalErr, "errs", errs)
		return finalErr
	}
	return nil
}

// Subscribe API
type Header struct {
	*types.Header
	*rpc.Client
}

// SubscribeNewHead subscribes to notifications about the current blockchain head
// on the given channel.
func (mc *Client) SubscribeNewHead(ctx context.Context, ch chan<- *Header) (ethereum.Subscription, error) {
	ids := mc.rpcClientMap.Ids()
	lens := len(ids)
	if lens == 0 {
		return nil, ErrNoEthClient
	}

	var subscribeNewHeadWg sync.WaitGroup

	cctx, cancel := context.WithCancel(ctx)
	for _, id := range ids {
		subscribeNewHeadWg.Add(1)
		go mc.subscribeNewHead(cctx, &subscribeNewHeadWg, id, ch)
	}

	newClientCh := mc.pubSub.Sub(newAvailableClientTopic)
	// handle new clients come
	go func() {
		defer mc.pubSub.Unsub(newClientCh, newAvailableClientTopic)
		for {
			select {
			case newC := <-newClientCh:
				id := newC.(uint64)
				subscribeNewHeadWg.Add(1)
				go mc.subscribeNewHead(cctx, &subscribeNewHeadWg, id, ch)
			case <-cctx.Done():
				return
			}
		}
	}()

	return event.NewSubscription(func(unsub <-chan struct{}) error {
		<-unsub
		cancel()
		subscribeNewHeadWg.Wait()
		return nil
	}), nil
}

func (mc *Client) subscribeNewHead(ctx context.Context, wg *sync.WaitGroup, id uint64, ch chan<- *Header) error {
	logger := log.New("id", id)
	defer wg.Done()

	retryTimer := time.NewTimer(0)
	defer retryTimer.Stop()

	for {
		url, rc := mc.rpcClientMap.GetById(id)
		if rc == nil {
			logger.Trace("EthClient has been removed")
			return nil
		}
		subLogger := logger.New("url", url)
		// If we have error, we need to retry
		err := doSubscribe(ctx, subLogger, rc, ch)
		if err == nil {
			return nil
		}

		// reset timer with retryPeriod
		retryTimer.Reset(retryPeriod)
		select {
		case <-retryTimer.C:
		case <-ctx.Done():
			return nil
		}
		subLogger.Trace("Retry to subscribe new head")
	}
}

func doSubscribe(ctx context.Context, logger log.Logger, rc *rpc.Client, ch chan<- *Header) error {
	headerCh := make(chan *types.Header)
	c := ethclient.NewClient(rc)
	subCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	sub, err := c.SubscribeNewHead(subCtx, headerCh)
	if err != nil {
		logger.Warn("Failed to subscribe new head", "err", err)
		return err
	}
	for {
		select {
		case header := <-headerCh:
			h := &Header{
				Client: rc,
				Header: header,
			}
			select {
			case ch <- h:
			case <-subCtx.Done():
				return nil
			}
		case err := <-sub.Err():
			logger.Warn("Failed during subscription", "err", err)
			return err
		case <-subCtx.Done():
			return nil
		}
	}
}

type dialedClient struct {
	url    string
	client *rpc.Client
}

func (mc *Client) DialClients(ctx context.Context) {
	urls := mc.rpcClientMap.NilClients()
	dialCh := make(chan *dialedClient, len(urls))
	for _, rawURL := range urls {
		go func(rawURL string) {
			dialCtx, cancel := context.WithTimeout(ctx, dialTimeout)
			defer cancel()
			c, err := rpc.DialContext(dialCtx, rawURL)
			if err == nil {
				log.Info("Connect to eth client successfully", "url", rawURL)
			} else {
				log.Warn("Failed to dial eth client", "url", rawURL, "err", err)
			}
			dialCh <- &dialedClient{url: rawURL, client: c}
		}(rawURL)
	}

	for i := 0; i < len(urls); i++ {
		dialed := <-dialCh
		if dialed.client != nil {
			id := mc.rpcClientMap.Replace(dialed.url, dialed.client)
			mc.pubSub.Pub(id, newAvailableClientTopic)
		}
	}
}

func (mc *Client) retrydial() {
	defer mc.retrydialWg.Done()

	ticker := time.NewTicker(retryPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-mc.newClientCh:
			mc.DialClients(mc.ctx)
		case <-ticker.C:
			mc.DialClients(mc.ctx)
		case <-mc.ctx.Done():
			// mc is closed, stop retry
			return
		}
	}
}
