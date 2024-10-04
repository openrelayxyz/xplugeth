package core

import (
	"github.com/openrelayxyz/xplugeth"
	"github.com/openrelayxyz/xplugeth/types"
	
	"github.com/ethereum/go-ethereum/node"
)


type Initializer interface {
	InitializeNode(*node.Node, types.Backend)
}
type Shutdown interface {
	Shutdown()
}
type Blockchain interface {
	Blockchain()	
}

func init() {
	xplugeth.RegisterHook[Initializer]()
	xplugeth.RegisterHook[Shutdown]()
	xplugeth.RegisterHook[Blockchain]()
}