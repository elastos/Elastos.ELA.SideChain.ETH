package dpos

import (
	"errors"
	"github.com/elastos/Elastos.ELA/events"
	"sync"

	"github.com/elastos/Elastos.ELA/common"
	"github.com/elastos/Elastos.ELA/core/types/payload"
)

const cachedCount = 6

type DBlock interface {
	Hash() common.Uint256
	Height() uint64
}

type ConfirmInfo struct {
	Confirm *payload.Confirm
	Height  uint64
}

type BlockPool struct {
	IsCurrent func() bool

	sync.RWMutex
	blocks   map[common.Uint256]DBlock
	confirms map[common.Uint256]*payload.Confirm

	OnConfirmBlock func(block DBlock, confirm *payload.Confirm) error
	VerifyConfirm  func(confirm *payload.Confirm) error
	VerifyBlock    func(block DBlock) error
}

func NewBlockPool(confirmBlock func(block DBlock, confirm *payload.Confirm) error,
	verifyConfirm func(confirm *payload.Confirm) error,
	verifyBlock func(block DBlock) error,
	isCurrent func() bool) *BlockPool {
	return &BlockPool{
		IsCurrent:      isCurrent,
		blocks:         make(map[common.Uint256]DBlock),
		confirms:       make(map[common.Uint256]*payload.Confirm),
		OnConfirmBlock: confirmBlock,
		VerifyConfirm:  verifyConfirm,
		VerifyBlock:    verifyBlock,
	}
}

func (bm *BlockPool) AppendConfirm(confirm *payload.Confirm) error {
	bm.Lock()
	defer bm.Unlock()

	return bm.appendConfirm(confirm)
}

func (bm *BlockPool) AppendDposBlock(dposBlock DBlock) error {
	bm.Lock()
	defer bm.Unlock()
	return bm.appendBlock(dposBlock)
}

func (bm *BlockPool) appendBlock(block DBlock) error {
	// add block
	hash := block.Hash()
	if _, ok := bm.blocks[hash]; ok {
		return errors.New("duplicate block in pool")
	}
	// verify block
	if err := bm.VerifyBlock(block); err != nil {
		Info("[AppendBlock] check block sanity failed, ", err)
		return err
	}

	bm.blocks[block.Hash()] = block

	// confirm block
	err := bm.confirmBlock(hash)
	if err != nil {
		Debug("[AppendDposBlock] ConfirmBlock failed, height", block.Height,
			"hash:", hash.String(), "err: ", err)
		return err
	}

	// notify new block received
	events.Notify(events.ETNewBlockReceived, block)

	return nil
}

func (bm *BlockPool) appendConfirm(confirm *payload.Confirm) error {

	// verify confirmation
	if err := bm.VerifyConfirm(confirm); err != nil {
		return err
	}
	bm.confirms[confirm.Proposal.BlockHash] = confirm

	err := bm.confirmBlock(confirm.Proposal.BlockHash)
	if err != nil {
		return err
	}
	block := bm.blocks[confirm.Proposal.BlockHash]

	// notify new confirm accepted.
	events.Notify(events.ETConfirmAccepted, &ConfirmInfo{
		Confirm: confirm,
		Height:  block.Height(),
	})

	return nil
}

func (bm *BlockPool) ConfirmBlock(hash common.Uint256) error {
	bm.Lock()
	err := bm.confirmBlock(hash)
	bm.Unlock()
	return err
}

func (bm *BlockPool) confirmBlock(hash common.Uint256) error {
	Info("[ConfirmBlock] block hash:", hash)

	block, ok := bm.blocks[hash]
	if !ok {
		return errors.New("there is no block in pool when confirming block")
	}

	confirm, ok := bm.confirms[hash]
	if !ok {
		return errors.New("there is no block confirmation in pool when confirming block")
	}

	if bm.OnConfirmBlock != nil {
		err := bm.OnConfirmBlock(block, confirm)
		if err != nil {
			return err
		}
	} else {
		panic("Not set OnConfirmBlock callBack")
	}

	return nil
}

func (bm *BlockPool) AddToBlockMap(block DBlock) {
	bm.Lock()
	defer bm.Unlock()

	bm.blocks[block.Hash()] = block
}

func (bm *BlockPool) GetBlock(hash common.Uint256) (DBlock, bool) {
	bm.RLock()
	defer bm.RUnlock()

	block, ok := bm.blocks[hash]
	return block, ok
}

func (bm *BlockPool) AddToConfirmMap(confirm *payload.Confirm) {
	bm.Lock()
	defer bm.Unlock()

	bm.confirms[confirm.Proposal.BlockHash] = confirm
}

func (bm *BlockPool) CleanFinalConfirmedBlock(height uint64) {
	bm.Lock()
	defer bm.Unlock()

	for _, block := range bm.blocks {
		if block.Height() < height-cachedCount {
			delete(bm.blocks, block.Hash())
			delete(bm.confirms, block.Hash())
		}
	}
}

func (bm *BlockPool) GetConfirm(hash common.Uint256) (*payload.Confirm, bool) {
	bm.Lock()
	defer bm.Unlock()

	confirm, ok := bm.confirms[hash]
	return confirm, ok
}
