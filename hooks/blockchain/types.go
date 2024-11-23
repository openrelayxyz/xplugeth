package blockchain

import (
	"math/big"
	"time"

	"github.com/openrelayxyz/xplugeth"
	
	"github.com/ethereum/go-ethereum/common"
	gtypes "github.com/ethereum/go-ethereum/core/types"
)

type NewHeadPlugin interface {
	NewHead(*gtypes.Block, common.Hash, []*gtypes.Log, *big.Int)
}

type NewSideBlockPlugin interface {
	NewSideBlock(*gtypes.Block, common.Hash, []*gtypes.Log)
}

type ReorgPlugin interface {
	Reorg(common.Hash, []common.Hash, []common.Hash)
}

type SetTrieFlushIntervalClonePlugin interface {
	SetTrieFlushIntervalClone(time.Duration) time.Duration
}

type PeerEvalPlugin interface {
	PeerEval(peerId string, headers []*gtypes.Header)
}


func init() {
	xplugeth.RegisterHook[NewHeadPlugin](blockchainPatchsets...)
	xplugeth.RegisterHook[NewSideBlockPlugin]()
	xplugeth.RegisterHook[ReorgPlugin]()
	xplugeth.RegisterHook[SetTrieFlushIntervalClonePlugin]()
	xplugeth.RegisterHook[PeerEvalPlugin]()
}
