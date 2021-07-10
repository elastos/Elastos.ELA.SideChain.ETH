// Copyright 2020 ChainSafe Systems
// SPDX-License-Identifier: LGPL-3.0-only

package evmclient

import (
	"encoding/json"
	"fmt"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/log"
	"os"
	"path/filepath"

	"github.com/elastos/Elastos.ELA.SideChain.ETH/chainbridge-core/config"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/chainbridge-core/crypto/secp256k1"
)

const DefaultGasLimit = 6721975
const DefaultGasPrice = 20000000000
const DefaultGasMultiplier = 1
const DefaultBlockConfirmations = 10
const DefaultConfigPath = "./chain_bridge.json"

type EVMConfig struct {
	SharedEVMConfig config.GeneralChainConfig
	kp              *secp256k1.Keypair
	EgsApiKey       string // API key for ethgasstation to query gas prices
	EgsSpeed        string // The speed which a transaction should be processed: average, fast, fastest. Default: fast
}

func NewConfig() *EVMConfig {
	return &EVMConfig{}
}

func GetConfig(path string) (*config.GeneralChainConfig, error) {
	var fig config.GeneralChainConfig
	if path == "" {
		path = DefaultConfigPath
	}

	err := loadConfig(path, &fig)
	if err != nil {
		log.Warn("err loading json file", "err", err.Error())
		return &fig, err
	}

	log.Debug("Loaded config", "path", path)
	err = fig.Validate()
	if err != nil {
		return nil, err
	}
	return &fig, nil

}

func ParseConfig(rawConfig *config.GeneralChainConfig) (*EVMConfig, error) {
	config := &EVMConfig{
		SharedEVMConfig: *rawConfig,
		EgsApiKey:       "",
		EgsSpeed:        "",
	}

	return config, nil
}

func loadConfig(file string, config *config.GeneralChainConfig) error {
	ext := filepath.Ext(file)
	fp, err := filepath.Abs(file)
	if err != nil {
		return err
	}
	log.Debug("Loading configuration", "path", filepath.Clean(fp))

	f, err := os.Open(filepath.Clean(fp))
	if err != nil {
		return err
	}

	if ext == ".json" {
		if err = json.NewDecoder(f).Decode(&config); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("unrecognized extention: %s", ext)
	}

	return nil
}