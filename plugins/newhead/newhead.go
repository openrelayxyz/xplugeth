package newhead

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	gethType "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/openrelayxyz/xplugeth"
	"github.com/openrelayxyz/xplugeth/types"
)

type newHeadModule struct{}

var (
	stack node.Node
)

func init() {
	xplugeth.RegisterModule[newHeadModule]()
}

func (*newHeadModule) InitializeNode(s *node.Node, b types.Backend) {
	stack = *s
	log.Info("new head module initialized")
}

func (*newHeadModule) PluginNewHead(block *gethType.Block, hash common.Hash, logs []*gethType.Log, td *big.Int) {
	log.Error("inside of the PluginNewHead")
	log.Info(
		"newHead", "block", block,
		"hash", hash,
		"logs", logs,
		"td", td)
}

func (*newHeadModule) PluginNewSideBlock(block *gethType.Block, hash common.Hash, logs []*gethType.Log) {
	log.Error("inside of the Plugin new side block")
	log.Info(
		"newSideBlock", "block", block,
		"hash", hash,
		"logs", logs)
}
