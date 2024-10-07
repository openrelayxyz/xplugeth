package initialize

import (
	"github.com/openrelayxyz/xplugeth"
	"github.com/openrelayxyz/xplugeth/types"
	
	"github.com/ethereum/go-ethereum/node"
)

type Shutdown interface {
	Shutdown()
}
type Blockchain interface {
	Blockchain()	
}

type Initializer interface {
	InitializeNode(*node.Node, types.Backend)
}

func init() {
	xplugeth.RegisterHook[Initializer]()
	xplugeth.RegisterHook[Shutdown]()
	xplugeth.RegisterHook[Blockchain]()
}