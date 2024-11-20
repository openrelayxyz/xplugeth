package blockchain

import (
	"github.com/openrelayxyz/xplugeth"
)

var blockchainPatchsets = []xplugeth.Patchset {
	xplugeth.Patchset{
		Remote: "github.com/openrelayxyz/xplugeth-patches",
		Ref: "hooks_foundation_blockchain_v1.14.12_0",
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