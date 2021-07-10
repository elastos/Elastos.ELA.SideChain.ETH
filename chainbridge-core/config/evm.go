// Copyright 2020 ChainSafe Systems
// SPDX-License-Identifier: LGPL-3.0-only

package config

import (
	"fmt"
)

const DefaultGasLimit = 6721975
const DefaultGasPrice = 20000000000
const DefaultGasMultiplier = 1
const DefaultBlockConfirmations = 10


type OpsConfig struct {
	Bridge             string  `mapstructure:"bridge"`
	WEthHandler       string   `mapstructure:"wethHandler"`
	Erc20Handler       string  `mapstructure:"erc20Handler"`
	Erc721Handler      string  `mapstructure:"erc721Handler"`
	GenericHandler     string  `mapstructure:"genericHandler"`
	MaxGasPrice        int64   `mapstructure:"maxGasPrice"`
	GasMultiplier      float64 `mapstructure:"gasMultiplier"`
	GasLimit           int64   `mapstructure:"gasLimit"`
	StartBlock         int64   `mapstructure:"startBlock"`
	BlockConfirmations int64   `mapstructure:"blockConfirmations"`
}

func (c *OpsConfig) Validate() error {
	if c.Bridge == "" {
		return fmt.Errorf("required field chain.Bridge empty for chain")
	}
	return nil
}

func (c *OpsConfig) ParseConfig() (*OpsConfig, error) {
	config := &OpsConfig{
		Erc20Handler:       c.Erc20Handler,
		Erc721Handler:      c.Erc721Handler,
		GenericHandler:     c.GenericHandler,
	}

	if c.Bridge != "" {
		config.Bridge = c.Bridge
	} else {
		return nil, fmt.Errorf("must provide opts.bridge field for ethereum config")
	}

	if c.GasLimit != 0 {
		config.GasLimit = c.GasLimit
	} else {
		config.GasLimit = DefaultGasLimit
	}

	if c.MaxGasPrice != 0 {
		config.MaxGasPrice = c.MaxGasPrice
	} else {
		config.MaxGasPrice = DefaultGasPrice
	}

	if c.GasMultiplier != 0 {
		config.GasMultiplier = c.GasMultiplier
	} else {
		config.GasMultiplier = DefaultGasMultiplier
	}

	if c.BlockConfirmations != 0 {
		config.BlockConfirmations = c.BlockConfirmations
	} else {
		config.BlockConfirmations = DefaultBlockConfirmations
	}

	config.StartBlock = c.StartBlock

	return config, nil
}
