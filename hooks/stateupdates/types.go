package stateupdates

import (
	"github.com/openrelayxyz/xplugeth"
	
	"github.com/ethereum/go-ethereum/common"
)

type StateUpdatePlugin interface {
	StateUpdate(common.Hash, common.Hash, map[common.Hash]struct{}, map[common.Hash][]byte, map[common.Hash]map[common.Hash][]byte, map[common.Hash][]byte)
}

func init() {
	xplugeth.RegisterHook[StateUpdatePlugin]()

}