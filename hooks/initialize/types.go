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
	xplugeth.RegisterHook[Initializer](
		xplugeth.Patchset{
			Remote: "github.com/openrelayxyz/xplugeth-patches",
			Ref: "hooks_etc_init_v1.12.20_2",
			Tests: []xplugeth.Test{
				{
					Package: "./cmd/geth",
					TestNames: []string{
						"TestGethPkgInjections",
					},
				},
			},
		},
	)
	xplugeth.RegisterHook[Shutdown]()
	xplugeth.RegisterHook[Blockchain]()
}