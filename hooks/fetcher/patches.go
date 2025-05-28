package fetcher

import (
	"github.com/openrelayxyz/xplugeth"
)

var fetcherPatchsets = []xplugeth.Patchset{
	xplugeth.Patchset{
		Remote: "github.com/openrelayxyz/xplugeth-patches",
		Ref:    "hooks_etc_fetcher_v1.12.20_2",
		// https://github.com/openrelayxyz/xplugeth-patches/commit/929e355704f59ef4b76871d4026692ac747b011a
		Tests: []xplugeth.Test{
			{
				Package: "./eth/fetcher",
				TestNames: []string{
					"TestFetcherPkgInjections",
				},
			},
		},
	},
	xplugeth.Patchset{
		Remote: "github.com/openrelayxyz/xplugeth-patches",
		Ref:    "hooks_bor_fetcher_v1.5.3_1",
		// https://github.com/openrelayxyz/xplugeth-patches/commit/ae076abdae9585c66a1a22a9dda7b9a8b5780be9
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