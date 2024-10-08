package blockchain

import (
	"math/big"
	"time"

	"github.com/openrelayxyz/xplugeth"
	
	"github.com/ethereum/go-ethereum/common"
	gtypes "github.com/ethereum/go-ethereum/core/types"
)

type NewHeadPlugin interface {
	NewHead(block *gtypes.Block, hash common.Hash, logs []*gtypes.Log, td *big.Int)
}

type NewSideBlockPlugin interface {
	NewSideBlock(block *gtypes.Block, hash common.Hash, logs []*gtypes.Log)
}

type ReorgPlugin interface {
	Reorg(commonBlock common.Hash, oldChain, newChain []common.Hash)
}

type SetTrieFlushIntervalClonePlugin interface {
	SetTrieFlushIntervalClone(flushInterval time.Duration) time.Duration
}

func init() {
	xplugeth.RegisterHook[NewHeadPlugin]()
	xplugeth.RegisterHook[NewSideBlockPlugin]()
	xplugeth.RegisterHook[ReorgPlugin]()
	xplugeth.RegisterHook[SetTrieFlushIntervalClonePlugin]()
}