package triecommit

import (
	"github.com/openrelayxyz/xplugeth"
)

var triecommitPatchsets = []xplugeth.Patchset {
	xplugeth.Patchset{
		Remote: "github.com/openrelayxyz/xplugeth-patches",
		Ref: "hooks_foundation_triecommit_v1.14.11_0",
		// https://github.com/openrelayxyz/xplugeth-patches/commit/250617992aa80c1b2170704a7fe177dcf2fb9ab5
		Tests: []xplugeth.Test{
			{
				Package: "./core",
				TestNames: []string{
					"TestHashDBInjections",
				},
			},
		},
	},
}