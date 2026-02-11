package initialize

import (
	"github.com/openrelayxyz/xplugeth"
)

var initializePatchsets = []xplugeth.Patchset {
	xplugeth.Patchset{
		Remote: "github.com/openrelayxyz/xplugeth-patches",
		Ref: "hooks_foundation_init_v1.16.5_2",
		// https://github.com/openrelayxyz/xplugeth-patches/commit/2cd24b3eb2890e62799d6a45d1a36f56060ad5aa
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
