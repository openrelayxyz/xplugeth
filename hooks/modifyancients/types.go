package modifyancients

import (
	"github.com/openrelayxyz/xplugeth"

	"github.com/ethereum/go-ethereum/core/types"
)

type ModifyAncientsPlugin interface {
	ModifyAncients(uint64, *types.Header)
}

func init() {
	xplugeth.RegisterHook[ModifyAncientsPlugin](
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
	)
}
