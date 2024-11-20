package modifyancients

import (
	"github.com/openrelayxyz/xplugeth"
)

var modifyancientsPatchsets = []xplugeth.Patchset {
	xplugeth.Patchset{
		Remote: "github.com/openrelayxyz/xplugeth-patches",
		Ref: "hooks_foundation_modifyancients_v1.14.11_0",
		Tests: []xplugeth.Test{
			{
				Package: "./core/rawdb",
				TestNames: []string{
					"TestRawDBInjections",
				},
			},
		},
	},
}