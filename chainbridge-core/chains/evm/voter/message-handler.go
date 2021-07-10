// Copyright 2020 ChainSafe Systems
// SPDX-License-Identifier: LGPL-3.0-only

package voter

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/log"
	"math/big"
	"strings"

	"github.com/elastos/Elastos.ELA.SideChain.ETH"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/accounts/abi"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/chainbridge-core/relayer"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/common"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/common/hexutil"
)

type MessageHandlerFunc func(m *relayer.Message, handlerAddr, bridgeAddress common.Address) (Proposer, error)

func NewEVMMessageHandler(client ChainClient, bridgeAddress common.Address) *EVMMessageHandler {
	return &EVMMessageHandler{
		bridgeAddress: bridgeAddress,
		client:        client,
	}
}

type EVMMessageHandler struct {
	client        ChainClient
	handlers      map[common.Address]MessageHandlerFunc
	bridgeAddress common.Address
}

func (mh *EVMMessageHandler) HandleMessage(m *relayer.Message) (Proposer, error) {
	// Matching resource ID with handler.
	addr, err := mh.matchResourceIDToHandlerAddress(m.ResourceId)
	// Based on handler that registered on BridgeContract
	log.Info("Handling new message", "type", m.Type, "src", m.Source, "dst", m.Destination, "nonce", m.DepositNonce, "rid", m.ResourceId, "handlerAddress", addr.String())
	handleMessage, err := mh.MatchAddressWithHandlerFunc(addr)
	if err != nil {
		return nil, err
	}

	prop, err := handleMessage(m, addr, mh.bridgeAddress)
	if err != nil {
		return nil, err
	}
	return prop, nil
}

func (mh *EVMMessageHandler) matchResourceIDToHandlerAddress(rID [32]byte) (common.Address, error) {
	definition := "[{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"_resourceIDToHandlerAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]"
	a, err := abi.JSON(strings.NewReader(definition))
	if err != nil {
		return common.Address{}, err
	}
	input, err := a.Pack("_resourceIDToHandlerAddress", rID)
	if err != nil {
		return common.Address{}, err
	}
	msg := ethereum.CallMsg{From: common.Address{}, To: &mh.bridgeAddress, Data: input}
	out, err := mh.client.CallContract(context.TODO(), toCallArg(msg), nil)
	if err != nil {
		return common.Address{}, err
	}
	out0 := common.Address{}
	err = a.Unpack(&out0, "_resourceIDToHandlerAddress", out)
	if err != nil {
		return common.Address{}, errors.New("no handler associated with such resourceID")
	}
	return out0, nil
}

func (mh *EVMMessageHandler) MatchAddressWithHandlerFunc(addr common.Address) (MessageHandlerFunc, error) {
	h, ok := mh.handlers[addr]
	if !ok {
		return nil, errors.New(fmt.Sprintf("no corresponding message handler for this address %s exists", addr.Hex()))
	}
	return h, nil
}

func (mh *EVMMessageHandler) RegisterMessageHandler(address common.Address, handler MessageHandlerFunc) {
	if mh.handlers == nil {
		mh.handlers = make(map[common.Address]MessageHandlerFunc)
	}
	mh.handlers[address] = handler
}

func ERC20MessageHandler(m *relayer.Message, handlerAddr, bridgeAddress common.Address) (Proposer, error) {
	if len(m.Payload) != 2 {
		return nil, errors.New("malformed payload. Len  of payload should be 2")
	}
	amount, ok := m.Payload[0].([]byte)
	if !ok {
		return nil, errors.New("wrong payloads amount format")
	}

	recipient, ok := m.Payload[1].([]byte)
	if !ok {
		return nil, errors.New("wrong payloads recipient format")

	}
	var data []byte
	data = append(data, common.LeftPadBytes(amount, 32)...) // amount (uint256)

	recipientLen := big.NewInt(int64(len(recipient))).Bytes()
	data = append(data, common.LeftPadBytes(recipientLen, 32)...) // length of recipient (uint256)
	data = append(data, recipient...)                             // recipient ([]byte)

	return &Proposal{
		Source:         m.Source,
		DepositNonce:   m.DepositNonce,
		ResourceId:     m.ResourceId,
		Data:           data,
		HandlerAddress: handlerAddr,
		BridgeAddress:  bridgeAddress,
	}, nil
}

func ERC721MessageHandler(msg *relayer.Message, handlerAddr, bridgeAddress common.Address) (*Proposal, error) {
	if len(msg.Payload) != 3 {
		return nil, errors.New("malformed payload. Len  of payload should be 3")
	}
	tokenID, ok := msg.Payload[0].([]byte)
	if !ok {
		return nil, errors.New("wrong payloads tokenID format")
	}
	recipient, ok := msg.Payload[1].([]byte)
	if !ok {
		return nil, errors.New("wrong payloads recipient format")
	}
	metadata, ok := msg.Payload[2].([]byte)
	if !ok {
		return nil, errors.New("wrong payloads metadata format")
	}

	data := bytes.Buffer{}
	data.Write(common.LeftPadBytes(tokenID, 32))

	recipientLen := big.NewInt(int64(len(recipient))).Bytes()
	data.Write(common.LeftPadBytes(recipientLen, 32))
	data.Write(recipient)

	metadataLen := big.NewInt(int64(len(metadata))).Bytes()
	data.Write(common.LeftPadBytes(metadataLen, 32))
	data.Write(metadata)
	return &Proposal{
		Source:         msg.Source,
		DepositNonce:   msg.DepositNonce,
		ResourceId:     msg.ResourceId,
		Data:           data.Bytes(),
		HandlerAddress: handlerAddr,
		BridgeAddress:  bridgeAddress,
	}, nil
}

func GenericMessageHandler(msg *relayer.Message, handlerAddr, bridgeAddress common.Address) (*Proposal, error) {
	if len(msg.Payload) != 1 {
		return nil, errors.New("malformed payload. Len  of payload should be 1")
	}
	metadata, ok := msg.Payload[0].([]byte)
	if !ok {
		return nil, errors.New("unable to convert metadata to []byte")
	}
	data := bytes.Buffer{}
	metadataLen := big.NewInt(int64(len(metadata))).Bytes()
	data.Write(common.LeftPadBytes(metadataLen, 32)) // length of metadata (uint256)
	data.Write(metadata)
	return &Proposal{
		Source:         msg.Source,
		DepositNonce:   msg.DepositNonce,
		ResourceId:     msg.ResourceId,
		Data:           data.Bytes(),
		HandlerAddress: handlerAddr,
		BridgeAddress:  bridgeAddress,
	}, nil
}

func toCallArg(msg ethereum.CallMsg) map[string]interface{} {
	arg := map[string]interface{}{
		"from": msg.From,
		"to":   msg.To,
	}
	if len(msg.Data) > 0 {
		arg["data"] = hexutil.Bytes(msg.Data)
	}
	if msg.Value != nil {
		arg["value"] = (*hexutil.Big)(msg.Value)
	}
	if msg.Gas != 0 {
		arg["gas"] = hexutil.Uint64(msg.Gas)
	}
	if msg.GasPrice != nil {
		arg["gasPrice"] = (*hexutil.Big)(msg.GasPrice)
	}
	return arg
}
