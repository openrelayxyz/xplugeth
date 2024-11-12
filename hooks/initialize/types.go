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
			Ref: "hooks_bor_init_v1.5.2_1",
			Tests: []xplugeth.Test{
				{
					Package: "./internal/cli/server",
					TestNames: []string{
						"TestServerPkgInjections",
					},
				},
			},
		},
	)
	xplugeth.RegisterHook[Shutdown]()
	xplugeth.RegisterHook[Blockchain]()
}