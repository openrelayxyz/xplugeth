package initialize

import (
	"github.com/openrelayxyz/xplugeth"
)

var initializePatchsets = []xplugeth.Patchset {
	xplugeth.Patchset{
		Remote: "github.com/openrelayxyz/xplugeth-patches",
		Ref: "hooks_foundation_init_v1.15.0_0",
		// https://github.com/openrelayxyz/xplugeth-patches/commit/be5ea0b65724131135ffe17971a1fdea2e41359d
		Tests: []xplugeth.Test{
			{
				Package: "./cmd/geth",
				TestNames: []string{
					"TestGethPkgInjections",
				},
			},
		},
	},
	xplugeth.Patchset{
		Remote: "github.com/openrelayxyz/xplugeth-patches",
		Ref: "hooks_foundation_init_v1.14.12_3",
		// https://github.com/openrelayxyz/xplugeth-patches/commit/8efabfea259b1209e00eff19e4fbac8c14f28953
		Tests: []xplugeth.Test{
			{
				Package: "./cmd/geth",
				TestNames: []string{
					"TestGethPkgInjections",
				},
			},
		},
	},
	xplugeth.Patchset{
		Remote: "github.com/openrelayxyz/xplugeth-patches",
		Ref: "hooks_etc_init_v1.12.20_5",
		// https://github.com/openrelayxyz/xplugeth-patches/commit/3519fd1ad995e10e0055e7d4d3d13cb54c24038a
		Tests: []xplugeth.Test{
			{
				Package: "./cmd/geth",
				TestNames: []string{
					"TestGethPkgInjections",
				},
			},
		},
	},
	xplugeth.Patchset{
		Remote: "github.com/openrelayxyz/xplugeth-patches",
		Ref: "hooks_bor_init_v1.5.3_1",
		// https://github.com/openrelayxyz/xplugeth-patches/commit/dc9954e395009c5f0a0c96eb18c1222dad7275c5
		Tests: []xplugeth.Test{
			{
				Package: "./internal/cli/server",
				TestNames: []string{
					"TestServerPkgInjections",
				},
			},
		},
	},
}