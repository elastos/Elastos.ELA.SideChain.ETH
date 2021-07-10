// Copyright 2020 ChainSafe Systems
// SPDX-License-Identifier: LGPL-3.0-only

package evmtransaction

import (
	"crypto/ecdsa"
	"math/big"

	"github.com/elastos/Elastos.ELA.SideChain.ETH/accounts/abi/bind"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/common"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/core/types"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/crypto"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/rlp"
)

type TX struct {
	tx *types.Transaction
}

// RawWithSignature mostly copies WithSignature interface of type.Transaction from go-ethereum,
// but return raw byte representation of transaction to be compatible and interchangeable between different go-ethereum forks
// WithSignature returns a new transaction with the given signature.
// This signature needs to be in the [R || S || V] format where V is 0 or 1.
func (a *TX) RawWithSignature(key *ecdsa.PrivateKey, chainID *big.Int) ([]byte, error) {
	opts := bind.NewKeyedTransactor(key)
	signer := types.NewEIP155Signer(chainID)
	tx, err := opts.Signer(signer, crypto.PubkeyToAddress(key.PublicKey), a.tx)
	if err != nil {
		return nil, err
	}
	a.tx = tx
	rawTX, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return nil, err
	}
	return rawTX, nil
}

func NewTransaction(nonce uint64, to common.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) *TX {
	tx := types.NewTransaction(nonce, to, amount, gasLimit, gasPrice, data)
	return &TX{tx: tx}
}

func (a *TX) Hash() common.Hash {
	return a.tx.Hash()
}
