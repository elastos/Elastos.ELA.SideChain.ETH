// Copyright 2020 ChainSafe Systems
// SPDX-License-Identifier: LGPL-3.0-only

package config

import (
	"fmt"
)

const DefaultKeystorePath = "./keys"
const DefaultBlockStore = "./blockStore"

var (
	DefaultConfigDir      = "./chainbridge/config/"
	KeystoreFlagName    = "./data/keystore/"
	BlockstoreFlagName  = "./chainbridge/blockstore"
	FreshStartFlagName  = "fresh"
	LatestBlockFlagName = "latest"
)

type GeneralChainConfig struct {
	Name           string `json:"name"` // Human-readable chain name
	Id             uint8 `json:"id"`   //ChainID
	Endpoint       string `json:"endpoint"`// url for rpc endpoint
	From           string `json:"from"`  // address of key to use
	KeystorePath   string // Location of key files
	Insecure       bool // Indicated whether the test keyring should be used
	BlockstorePath string // Location of blockstore
	FreshStart     bool  // If true, blockstore is ignored at start.
	LatestBlock    bool // If true, overrides blockstore or latest block in config and starts from current block
	Opts          OpsConfig
}

func (c *GeneralChainConfig) Validate() error {
	// viper defaults to 0 for not specified ints, but we must have a valid chain id
	// Previous method of checking used a string cast like below
	//chainId := string(c.Id)
	if c.Id == 0 {
		return fmt.Errorf("required field chain.Id empty for chain %v", c.Id)
	}
	if c.Endpoint == "" {
		return fmt.Errorf("required field chain.Endpoint empty for chain %v", c.Id)
	}
	if c.Name == "" {
		return fmt.Errorf("required field chain.Name empty for chain %v", c.Id)
	}
	if c.From == "" {
		return fmt.Errorf("required field chain.From empty for chain %v", c.Id)
	}
	return nil
}

func (c *GeneralChainConfig) ParseConfig() {
	//c.KeystorePath = DefaultKeystorePath
	//c.Insecure = false
	//c.BlockstorePath = DefaultBlockStore
	//c.FreshStart = true
	//c.LatestBlock = true
}


