package example

import (
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"

	"github.com/openrelayxyz/xplugeth"
	"github.com/openrelayxyz/xplugeth/types"
)

var stack node.Node

type demoModule struct{}

func (*demoModule) InitializeNode(s *node.Node, b types.Backend) {
	stack = *s
	log.Info("stack demo module initialized")
}

func init() {
	xplugeth.RegisterModule[demoModule]()
}
