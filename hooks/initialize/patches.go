package initialize

import (
	"github.com/openrelayxyz/xplugeth"
)

var initializePatchsets = []xplugeth.Patchset {
	xplugeth.Patchset{
		Remote: "github.com/openrelayxyz/xplugeth-patches",
		Ref: "hooks_foundation_init_v1.14.12_2",
		// https://github.com/openrelayxyz/xplugeth-patches/commit/f0afe27cd27e85342d03f7fa96f6d1776c6422f4
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
		Ref: "hooks_etc_init_v1.12.20_4",
		// https://github.com/openrelayxyz/xplugeth-patches/commit/4c6661805f26b47781321b92c703fe49a00c9c2f
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