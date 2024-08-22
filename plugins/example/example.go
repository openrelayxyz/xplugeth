package example

import (
	"github.com/openrelayxyz/xplugeth"
	"github.com/openrelayxyz/xplugeth/types"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/log"
)

var (
	sessionStack *node.Node
	sessionBackend types.Backend
)

type exampleModule struct {}

func (*exampleModule) InitializeNode(s *node.Node, b types.Backend) {
	log.Info("Example module initialized")
}

func (*exampleModule) Shutdown() {
	log.Info("Byeee!")
}

func init() {
	xplugeth.RegisterModule[exampleModule]()
}