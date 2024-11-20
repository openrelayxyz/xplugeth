package stateupdates

import (
	"github.com/openrelayxyz/xplugeth"
)

var stateupdatesPatchsets = []xplugeth.Patchset {
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
}