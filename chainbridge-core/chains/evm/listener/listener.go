// Copyright 2021 ChainSafe Systems
// SPDX-License-Identifier: LGPL-3.0-only

package listener

import (
	"context"

	"math/big"
	"time"

	"github.com/elastos/Elastos.ELA.SideChain.ETH/chainbridge-core/blockstore"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/chainbridge-core/relayer"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/common"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/log"
)

var BlockRetryInterval = time.Second * 5
var BlockDelay = big.NewInt(10) //TODO: move to config

type DepositLogs struct {
	DestinationID uint8
	ResourceID    [32]byte
	DepositNonce  uint64
}

type ChainClient interface {
	LatestBlock() (*big.Int, error)
	FetchDepositLogs(ctx context.Context, address common.Address, startBlock *big.Int, endBlock *big.Int) ([]*DepositLogs, error)
	CallContract(ctx context.Context, callArgs map[string]interface{}, blockNumber *big.Int) ([]byte, error)
}

type EventHandler interface {
	HandleEvent(sourceID, destID uint8, nonce uint64, rID [32]byte) (*relayer.Message, error)
}

type EVMListener struct {
	chainReader   ChainClient
	eventHandler  EventHandler
	bridgeAddress common.Address
}

func NewEVMListener(chainReader ChainClient, handler EventHandler, bridgeAddress common.Address) *EVMListener {
	return &EVMListener{chainReader: chainReader, eventHandler: handler, bridgeAddress: bridgeAddress}
}

func (l *EVMListener) ListenToEvents(startBlock *big.Int, chainID uint8, kvrw blockstore.KeyValueWriter, stopChn <-chan struct{}, errChn chan<- error) <-chan *relayer.Message {
	// TODO: This channel should be closed somewhere!
	ch := make(chan *relayer.Message)
	go func() {
		for {
			select {
			case <-stopChn:
				return
			default:
				head, err := l.chainReader.LatestBlock()
				if err != nil {
					time.Sleep(BlockRetryInterval)
					continue
				}
				// Sleep if the difference is less than BlockDelay; (latest - current) < BlockDelay
				if big.NewInt(0).Sub(head, startBlock).Cmp(BlockDelay) == -1 {
					time.Sleep(BlockRetryInterval)
					continue
				}
				logs, err := l.chainReader.FetchDepositLogs(context.Background(), l.bridgeAddress, startBlock, startBlock)
				if err != nil {
					// Filtering logs error really can appear only on wrong configuration or temporary network problem
					// so i do no see any reason to break execution
					log.Error("FetchDepositLogs errors", "ChainID", chainID,)
					continue
				}
				for _, eventLog := range logs {
					m, err := l.eventHandler.HandleEvent(chainID, eventLog.DestinationID, eventLog.DepositNonce, eventLog.ResourceID)
					if err != nil {
						errChn <- err
						log.Error("HandleEvent error", "error", err)
						return
					}
					log.Debug("Resolved message %+v in block %s", m, startBlock.String())
					ch <- m
				}
				if startBlock.Int64()%20 == 0 {
					// Logging process every 20 bocks to exclude spam
					log.Debug("Queried block for deposit events", "block", startBlock.String(), "chainId", chainID)
				}
				// TODO: We can store blocks to DB inside listener or make listener send something to channel each block to save it.
				//Write to block store. Not a critical operation, no need to retry
				err = blockstore.StoreBlock(kvrw, startBlock, chainID)
				if err != nil {
					log.Error("Failed to write latest block to blockstore", "block", startBlock.String())
				}
				// Goto next block
				startBlock.Add(startBlock, big.NewInt(1))
			}
		}
	}()
	return ch
}
