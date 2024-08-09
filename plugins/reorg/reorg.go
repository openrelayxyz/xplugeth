package reorg

import (
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/openrelayxyz/xplugeth"
	"github.com/openrelayxyz/xplugeth/shared"
	"github.com/openrelayxyz/xplugeth/types"
)

type reorgModule struct{}

var stack node.Node

func init() {
	xplugeth.RegisterModule[reorgModule]()
}

func (*reorgModule) InitializeNode(s *node.Node, b types.Backend) {
	stack = *s
	log.Error("inside of newOrg")
}

func (*reorgModule) Blockchain() {
	shared.Blockchain(stack)
}

// func (*reorgModule) PluginReorg(commonBlock *gethType.Block, oldChain, newChain gethType.Blocks) {

// }

func (*reorgModule) PluginSetTrieFlushIntervalClone(flushInterval time.Duration) time.Duration {
	return flushInterval
}
