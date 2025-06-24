package initialize

import (
	"github.com/openrelayxyz/xplugeth"
)

var initializePatchsets = []xplugeth.Patchset {
	xplugeth.Patchset{
		Remote: "github.com/openrelayxyz/xplugeth-patches",
		Ref: "hooks_foundation_init_v1.15.0_0",
		// https://github.com/openrelayxyz/xplugeth-patches/commit/be5ea0b65724131135ffe17971a1fdea2e41359d
		Tests: []xplugeth.Test{
			{
				Package: "./cmd/geth",
				TestNames: []string{
					"TestGethPkgInjections",
				},
			},
		},
	},
	xplugeth.Patchset{
		Remote: "github.com/openrelayxyz/xplugeth-patches",
		Ref: "hooks_foundation_init_v1.14.12_4",
		// https://github.com/openrelayxyz/xplugeth-patches/commit/51cc5dd64eb7fbeb72d47613da50778b52601ef4
		Tests: []xplugeth.Test{
			{
				Package: "./cmd/geth",
				TestNames: []string{
					"TestGethPkgInjections",
				},
			},
		},
	},
	xplugeth.Patchset{
		Remote: "github.com/openrelayxyz/xplugeth-patches",
		Ref: "hooks_etc_init_v1.12.20_6",
		// https://github.com/openrelayxyz/xplugeth-patches/commit/29c282fcf9f9d8276df87a343cbed510db6f8133
		Tests: []xplugeth.Test{
			{
				Package: "./cmd/geth",
				TestNames: []string{
					"TestGethPkgInjections",
				},
			},
		},
	},
	xplugeth.Patchset{
		Remote: "github.com/openrelayxyz/xplugeth-patches",
		Ref: "hooks_bor_init_v1.5.3_2",
		// https://github.com/openrelayxyz/xplugeth-patches/commit/b7184a28f74cfb56402cbf703a8246ea8831742a
		Tests: []xplugeth.Test{
			{
				Package: "./internal/cli/server",
				TestNames: []string{
					"TestServerPkgInjections",
				},
			},
		},
	},
	xplugeth.Patchset{
		Remote: "github.com/openrelayxyz/xplugeth-patches",
		Ref: "hooks_bor_init_v2.0.3_0",
		// https://github.com/openrelayxyz/xplugeth-patches/commit/c4704227fb138bc1194b98e8282d56fed128016c
		Tests: []xplugeth.Test{
			{
				Package: "./internal/cli/server",
				TestNames: []string{
					"TestServerPkgInjections",
				},
			},
		},
	},
}
