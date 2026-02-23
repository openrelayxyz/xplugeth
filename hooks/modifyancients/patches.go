package modifyancients

import (
	"github.com/openrelayxyz/xplugeth"
)

var modifyancientsPatchsets = []xplugeth.Patchset {
	xplugeth.Patchset{
		Remote: "github.com/openrelayxyz/xplugeth-patches",
		Ref: "hooks_foundation_modifyancients_v1.17.0_1",
		// https://github.com/openrelayxyz/xplugeth-patches/commit/3a071885f2de46c8786ce033914f3a3ab4f28ba7
		Tests: []xplugeth.Test{
			{
				Package: "./core/rawdb",
				TestNames: []string{
					"TestRawDBInjections",
				},
			},
		},
	},
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