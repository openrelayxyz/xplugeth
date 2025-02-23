package stateupdates

import (
	"github.com/openrelayxyz/xplugeth"
)

var stateupdatesPatchsets = []xplugeth.Patchset {
	xplugeth.Patchset{
		Remote: "github.com/openrelayxyz/xplugeth-patches",
		Ref: "hooks_foundation_stateupdates_v1.15.0_2",
		// https://github.com/openrelayxyz/xplugeth-patches/commit/8bf60ed0624851d21b94fb67c2ddec4a7ff56d91
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
		Ref: "hooks_foundation_stateupdates_v1.14.11_0",
		// https://github.com/openrelayxyz/xplugeth-patches/commit/49b20e469b3702109e01d7480c1ae757941823c6
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
		// https://github.com/openrelayxyz/xplugeth-patches/commit/67720ccd8d0d3c03992585ea359761622732eb75
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
		// https://github.com/openrelayxyz/xplugeth-patches/commit/edc23fd698afc8edf68e2636da3b41de37fa4739
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