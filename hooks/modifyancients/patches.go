package modifyancients

import (
	"github.com/openrelayxyz/xplugeth"
)

var modifyancientsPatchsets = []xplugeth.Patchset {
	xplugeth.Patchset{
		Remote: "github.com/openrelayxyz/xplugeth-patches",
		Ref: "hooks_foundation_modifyancients_v1.14.11_0",
		// https://github.com/openrelayxyz/xplugeth-patches/commit/77c776906566401407e9cf8fb7bade2b59b50376
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