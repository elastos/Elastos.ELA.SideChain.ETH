// Copyright 2020 ChainSafe Systems
// SPDX-License-Identifier: LGPL-3.0-only

package secp256sender

import (
	"crypto/ecdsa"

	"github.com/elastos/Elastos.ELA.SideChain.ETH/common"
)

type SecpInMemory256Sender struct {
	privateKey *ecdsa.PrivateKey
	address common.Address
}


func (s *SecpInMemory256Sender) PrivateKey() *ecdsa.PrivateKey {
	return s.privateKey
}
func (s *SecpInMemory256Sender) Address() string {
	return s.address.Hex()
}
