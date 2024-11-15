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
		xplugeth.Patchset{
			Remote: "github.com/openrelayxyz/xplugeth-patches",
			Ref: "hooks_etc_stateupdates_v1.12.20_1",
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