package fetcher

import (
	gtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/openrelayxyz/xplugeth"
)

type PeerEvalPlugin interface {
	PeerEval(peerId string, headers []*gtypes.Header)
}

func init() {
	xplugeth.RegisterHook[PeerEvalPlugin](fetcherPatchsets...)
}
