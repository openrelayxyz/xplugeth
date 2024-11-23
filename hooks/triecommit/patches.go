package triecommit

import (
	"github.com/openrelayxyz/xplugeth"
)

var triecommitPatchsets = []xplugeth.Patchset {
	xplugeth.Patchset{
		Remote: "github.com/openrelayxyz/xplugeth-patches",
		Ref: "hooks_foundation_triecommit_v1.14.11_0",
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