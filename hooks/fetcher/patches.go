package fetcher

import (
	"github.com/openrelayxyz/xplugeth"
)

var fetcherPatchsets = []xplugeth.Patchset{
	xplugeth.Patchset{
		Remote: "github.com/openrelayxyz/xplugeth-patches",
		Ref:    "hooks_bor_fetcher_v1.5.2_1",
		// https://github.com/openrelayxyz/xplugeth-patches/commit/bb974047d4988ada256711d421468b19d3556f01
		Tests: []xplugeth.Test{
			{
				Package: "./eth/fetcher",
				TestNames: []string{
					"TestFetcherPkgInjections",
				},
			},
		},
	},
}
