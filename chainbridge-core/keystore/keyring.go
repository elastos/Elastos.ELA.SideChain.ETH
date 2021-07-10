// Copyright 2020 ChainSafe Systems
// SPDX-License-Identifier: LGPL-3.0-only

package keystore

import (
	"fmt"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/chainbridge-core/crypto"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/chainbridge-core/crypto/secp256k1"
)

// The Constant "keys". These are the name that the keys are based on. This can be expanded, but
// any additions must be added to Keys and to insecureKeyFromAddress
const AliceKey = "alice"
const BobKey = "bob"
const CharlieKey = "charlie"
const DaveKey = "dave"
const EveKey = "eve"

var Keys = []string{AliceKey, BobKey, CharlieKey, DaveKey, EveKey}

// The Chain type Constants
const EthChain = "ethereum"
const SubChain = "substrate"

var TestKeyRing *TestKeyRingHolder

// TestKeyStore is a struct that holds a Keystore of all the test keys
type TestKeyRingHolder struct {
	EthereumKeys  map[string]*secp256k1.Keypair
	SubstrateKeys map[string]*secp256k1.Keypair
}

// Init function to create a keyRing that can be accessed anywhere without having to recreate the data
func init() {
	TestKeyRing = &TestKeyRingHolder{
		EthereumKeys: makeEthRing(),
	}
	TestKeyRing.SubstrateKeys = TestKeyRing.EthereumKeys
}

func makeEthRing() map[string]*secp256k1.Keypair {
	ring := map[string]*secp256k1.Keypair{}
	for _, key := range Keys {
		bz := padWithZeros([]byte(key), secp256k1.PrivateKeyLength)
		kp, err := secp256k1.NewKeypairFromPrivateKey(bz)
		if err != nil {
			panic(err)
		}
		ring[key] = kp
	}

	return ring
}

// padWithZeros adds on extra 0 bytes to make a byte array of a specified length
func padWithZeros(key []byte, targetLength int) []byte {
	res := make([]byte, targetLength-len(key))
	return append(res, key...)
}

// insecureKeypairFromAddress is used for resolving addresses to test keypairs.
func insecureKeypairFromAddress(key string, chainType string) (crypto.Keypair, error) {
	var kp crypto.Keypair
	var ok bool

	if chainType == EthChain {
		kp, ok = TestKeyRing.EthereumKeys[key]
	} else if chainType == SubChain {
		kp, ok = TestKeyRing.SubstrateKeys[key]
	} else {
		return nil, fmt.Errorf("unrecognized chain type: %s", chainType)
	}

	if !ok {
		return nil, fmt.Errorf("invalid test key selection: %s", key)
	}

	return kp, nil
}
