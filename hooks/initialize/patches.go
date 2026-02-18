package initialize

import (
	"github.com/openrelayxyz/xplugeth"
)

var initializePatchsets = []xplugeth.Patchset {
	xplugeth.Patchset{
		Remote: "github.com/openrelayxyz/xplugeth-patches",
		Ref: "hooks_foundation_init_v1.17.0_0",
		// https://github.com/openrelayxyz/xplugeth-patches/commit/2f3f2dd1b34267cde38ff46ea0bfbceb594aba1c
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
		Ref: "hooks_etc_init_v1.12.20_9",
		// https://github.com/openrelayxyz/xplugeth-patches/commit/bea9c77474bdb2958ace2e4f8ee7d862ac691248
		Tests: []xplugeth.Test{
			{
				Package: "./cmd/geth",
				TestNames: []string{
					"TestGethPkgInjections",
				},
			},
		},
	},
}
