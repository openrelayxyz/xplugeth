package modifyancients

import (
	"github.com/openrelayxyz/xplugeth"

	"github.com/ethereum/go-ethereum/core/types"
)

type ModifyAncientsPlugin interface {
	ModifyAncients(uint64, *types.Header)
}

func init() {
	xplugeth.RegisterHook[ModifyAncientsPlugin]()
}
