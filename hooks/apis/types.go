package apis

import (
	"github.com/openrelayxyz/xplugeth"
	"github.com/openrelayxyz/xplugeth/types"
	
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/rpc"
)

type GetAPIs interface {
	GetAPIs(*node.Node, types.Backend) []rpc.API
}

func init() {
	xplugeth.RegisterHook[GetAPIs]()
}