package blockchain

import (
	"github.com/openrelayxyz/xplugeth"
)

var blockchainPatchsets = []xplugeth.Patchset {
	xplugeth.Patchset{
		Remote: "github.com/openrelayxyz/xplugeth-patches",
		Ref: "hooks_foundation_blockchain_v1.15.0_1",
		// https://github.com/openrelayxyz/xplugeth-patches/commit/20867e93eb9f48d18e22c47b8e36bd545b9763cb
		Tests: []xplugeth.Test{
			{
				Package: "./core",
				TestNames: []string{
					"TestCoreInjections",
				},
			},
		},
	},
	xplugeth.Patchset{
		Remote: "github.com/openrelayxyz/xplugeth-patches",
		Ref: "hooks_foundation_blockchain_v1.14.12_0",
		// https://github.com/openrelayxyz/xplugeth-patches/commit/b1156a059f9ede71320144f7fbfae525fc60405c
		Tests: []xplugeth.Test{
			{
				Package: "./core",
				TestNames: []string{
					"TestCoreInjections",
				},
			},
		},
	},
	xplugeth.Patchset{
		Remote: "github.com/openrelayxyz/xplugeth-patches",
		Ref: "hooks_etc_blockchain_v1.12.20_0",
		// https://github.com/openrelayxyz/xplugeth-patches/commit/9054cb07cbc676582bd891595a5117edd1b8fb7e
		Tests: []xplugeth.Test{
			{
				Package: "./core",
				TestNames: []string{
					"TestCoreInjections",
				},
			},
		},
	},
	xplugeth.Patchset{
		Remote: "github.com/openrelayxyz/xplugeth-patches",
		Ref: "hooks_bor_blockchain_v2.1.0-beta3_0",
		// https://github.com/openrelayxyz/xplugeth-patches/commit/91bf0211e5d6f849324f36694392406bb5446b06
		Tests: []xplugeth.Test{
			{
				Package: "./core",
				TestNames: []string{
					"TestCoreInjections",
				},
			},
		},
	},
}