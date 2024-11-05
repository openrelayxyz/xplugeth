package stateupdates

import (
	"github.com/openrelayxyz/xplugeth"
	
	"github.com/ethereum/go-ethereum/common"
)

type StateUpdatePlugin interface {
	StateUpdate(blockRoot, parentRoot common.Hash, destructs map[common.Hash]struct{}, accounts map[common.Hash][]byte, storage map[common.Hash]map[common.Hash][]byte, codeUpdates map[common.Hash][]byte)
}

func init() {
	xplugeth.RegisterHook[StateUpdatePlugin](
		xplugeth.Patchset{
			Remote: "github.com/openrelayxyz/xplugeth-patches",
			Ref: "hooks_state_foundation_v1.14.11_0",
			Tests: []xplugeth.Test{
				{
					Package: "./core/state",
					TestNames: []string{
						"TestStateInjections",
					},
				},
			},
		},
	)
}