package fetcher

import (
	"github.com/openrelayxyz/xplugeth"
)

var fetcherPatchsets = []xplugeth.Patchset{
	xplugeth.Patchset{
		Remote: "github.com/openrelayxyz/xplugeth-patches",
		Ref:    "hooks_fetcher_bor_v1.5.2_0",
		// https://github.com/openrelayxyz/xplugeth-patches/commit/99e551b2dbe16a411b9e1ddc92f1e5bbafd1a7f2
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
