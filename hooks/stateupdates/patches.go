package stateupdates

import (
	"github.com/openrelayxyz/xplugeth"
)

var stateupdatesPatchsets = []xplugeth.Patchset {
	xplugeth.Patchset{
		Remote: "github.com/openrelayxyz/xplugeth-patches",
		Ref: "hooks_foundation_stateupdates_v1.15.6_0",
		// https://github.com/openrelayxyz/xplugeth-patches/commit/ef36f13af76f2c90b493330af509ed412b2b3342
		Tests: []xplugeth.Test{
			{
				Package: "./core/state",
				TestNames: []string{
					"TestStateInjections",
				},
			},
		},
	},
	xplugeth.Patchset{
		Remote: "github.com/openrelayxyz/xplugeth-patches",
		Ref: "hooks_etc_stateupdates_v1.12.20_2",
		// https://github.com/openrelayxyz/xplugeth-patches/commit/eb66450d205ecd961cb9db113f25e2b5769ddfdb
		Tests: []xplugeth.Test{
			{
				Package: "./core/state",
				TestNames: []string{
					"TestStateInjections",
				},
			},
		},
	},
	xplugeth.Patchset{
		Remote: "github.com/openrelayxyz/xplugeth-patches",
		Ref: "hooks_bor_stateupdates_v1.5.2_2",
		// https://github.com/openrelayxyz/xplugeth-patches/commit/ab3ea05e2f72c9daec8bcbd062185127a36a8793
		Tests: []xplugeth.Test{
			{
				Package: "./core/state",
				TestNames: []string{
					"TestStateInjections",
				},
			},
		},
	},
	xplugeth.Patchset{
		Remote: "github.com/openrelayxyz/xplugeth-patches",
		Ref: "hooks_bor_stateupdates_v2.0.3_0",
		// https://github.com/openrelayxyz/xplugeth-patches/commit/1662bf93cc20522a082d8e27832b6f8048b70638
		Tests: []xplugeth.Test{
			{
				Package: "./core/state",
				TestNames: []string{
					"TestStateInjections",
				},
			},
		},
	},
}