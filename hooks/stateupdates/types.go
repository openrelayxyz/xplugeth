package stateupdates

import (
	"github.com/openrelayxyz/xplugeth"
	
	"github.com/ethereum/go-ethereum/common"
)

type StateUpdatePlugin interface {
	StateUpdate(common.Hash, common.Hash, map[common.Hash]struct{}, map[common.Hash][]byte, map[common.Hash]map[common.Hash][]byte, map[common.Hash][]byte)
}

func init() {
	xplugeth.RegisterHook[StateUpdatePlugin](
		xplugeth.Patchset{
			Remote: "github.com/openrelayxyz/xplugeth-patches",
			Ref: "hooks_foundation_stateupdates_v1.14.11_0",
			Tests: []xplugeth.Test{
				{
					Package: "./core/state",
					TestNames: []string{
						"TestStateInjections",
					},
				},
			},
		},
		xplugeth.Patchset{
			Remote: "github.com/openrelayxyz/xplugeth-patches",
			Ref: "hooks_bor_stateupdates_v1.5.2_0",
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