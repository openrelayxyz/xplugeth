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

func init() {
	xplugeth.RegisterHook[NewHeadPlugin](
		xplugeth.Patchset{
			Remote: "github.com/openrelayxyz/xplugeth-patches",
			Ref: "hooks_foundation_blockchain_v1.14.11_0",
			Tests: []xplugeth.Test{
				{
					Package: "./core",
					TestNames: []string{
						"TestCoreInjections",
					},
				},
			},
		},
	)
	xplugeth.RegisterHook[NewSideBlockPlugin]()
	xplugeth.RegisterHook[ReorgPlugin]()
	xplugeth.RegisterHook[SetTrieFlushIntervalClonePlugin]()
}