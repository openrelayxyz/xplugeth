package example

import (
	"github.com/openrelayxyz/xplugeth"
	"github.com/openrelayxyz/xplugeth/types"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/log"
)


type exampleModule struct {}

func (*exampleModule) InitializeNode(s *node.Node, b types.Backend) {
	log.Info("Example module initialized", "s", s, "b", b)
}

func init() {
	xplugeth.RegisterModule[exampleModule]()
}