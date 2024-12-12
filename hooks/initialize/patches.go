package initialize

import (
	"github.com/openrelayxyz/xplugeth"
)

var initializePatchsets = []xplugeth.Patchset {
	xplugeth.Patchset{
		Remote: "github.com/openrelayxyz/xplugeth-patches",
		Ref: "hooks_foundation_init_v1.14.12_1",
		// https://github.com/openrelayxyz/xplugeth-patches/commit/9b3a39dbf7a09855123307329b042a3d1821e74b
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
		Ref: "hooks_etc_init_v1.12.20_3",
		// https://github.com/openrelayxyz/xplugeth-patches/commit/7adbb2753c0c2a2bc31ddcbd541f4f4be35a597a
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
		Ref: "hooks_bor_init_v1.5.3_0",
		// https://github.com/openrelayxyz/xplugeth-patches/commit/36cbeb7a405f28d628237c08772e3343522e9fa8
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