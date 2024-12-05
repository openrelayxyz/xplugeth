package blockchain

import (
	"github.com/openrelayxyz/xplugeth"
)

var blockchainPatchsets = []xplugeth.Patchset {
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
		Ref: "hooks_foundation_blockchain_v1.14.11_0",
		// https://github.com/openrelayxyz/xplugeth-patches/commit/a6d367ddb5a8b675e4eb67b8c69d52bae42c07d5
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
		Ref: "hooks_bor_blockchain_v1.5.2_2",
		// https://github.com/openrelayxyz/xplugeth-patches/commit/9a923faf4dea704d5d3a3bfbde30a81ba536001c
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