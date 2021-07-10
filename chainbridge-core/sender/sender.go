// Copyright 2020 ChainSafe Systems
// SPDX-License-Identifier: LGPL-3.0-only

package sender

import (
	"crypto/ecdsa"
)

type Sender interface {
	PrivateKey() *ecdsa.PrivateKey
	Address() string
}