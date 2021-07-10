// Copyright 2021 ChainSafe Systems
// SPDX-License-Identifier: LGPL-3.0-only

package evm

import (
	"fmt"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/log"
	"math/big"

	"github.com/elastos/Elastos.ELA.SideChain.ETH/chainbridge-core/blockstore"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/chainbridge-core/config"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/chainbridge-core/relayer"
)

type EventListener interface {
	ListenToEvents(startBlock *big.Int, chainID uint8, kvrw blockstore.KeyValueWriter, stopChn <-chan struct{}, errChn chan<- error) <-chan *relayer.Message
}

type ProposalVoter interface {
	VoteProposal(message *relayer.Message) error
}

// EVMChain is struct that aggregates all data required for
type EVMChain struct {
	listener              EventListener // Rename
	writer                ProposalVoter
	chainID               uint8
	kvdb                  blockstore.KeyValueReaderWriter
	bridgeContractAddress string
	config                *config.GeneralChainConfig
}

func NewEVMChain(dr EventListener, writer ProposalVoter, kvdb blockstore.KeyValueReaderWriter, chainID uint8, config *config.GeneralChainConfig) *EVMChain {
	return &EVMChain{listener: dr, writer: writer, kvdb: kvdb, chainID: chainID, config: config}
}

// PollEvents is the goroutine that polling blocks and searching Deposit Events in them. Event then sent to eventsChan
func (c *EVMChain) PollEvents(stop <-chan struct{}, sysErr chan<- error, eventsChan chan *relayer.Message) {
	log.Info("Polling Blocks...")
	// Handler chain specific configs and flags
	block, err := blockstore.SetupBlockstore(c.config, c.kvdb, big.NewInt(c.config.Opts.StartBlock))
	if err != nil {
		sysErr <- fmt.Errorf("error %w on getting last stored block", err)
		return
	}
	ech := c.listener.ListenToEvents(block, c.chainID, c.kvdb, stop, sysErr)
	for {
		select {
		case <-stop:
			return
		case newEvent := <-ech:
			// Here we can place middlewares for custom logic?
			eventsChan <- newEvent
			continue
		}
	}
}

func (c *EVMChain) Write(msg *relayer.Message) error {
	return c.writer.VoteProposal(msg)
}

func (c *EVMChain) ChainID() uint8 {
	return c.chainID
}
