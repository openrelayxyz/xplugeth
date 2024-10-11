package stateupdates

import (
	"github.com/openrelayxyz/xplugeth"
	
	"github.com/ethereum/go-ethereum/common"
)

type StateUpdatePlugin interface {
	StateUpdate(blockRoot, parentRoot common.Hash, destructs map[common.Hash]struct{}, accounts map[common.Hash][]byte, storage map[common.Hash]map[common.Hash][]byte, codeUpdates map[common.Hash][]byte)
}

func init() {
	xplugeth.RegisterHook[StateUpdatePlugin]()

}