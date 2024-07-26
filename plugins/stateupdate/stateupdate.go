package stateupdate

import (
	"github.com/openrelayxyz/xplugeth"

	"github.com/ethereum/go-ethereum/common"
)


type stateUpdateModule struct {}

var SUCount int

func (*stateUpdateModule) PluginStateUpdate(blockRoot, parentRoot common.Hash, destructs map[common.Hash]struct{}, accounts map[common.Hash][]byte, storage map[common.Hash]map[common.Hash][]byte, codeUpdates map[common.Hash][]byte) {
	SUCount += 1
}

func init() {
	xplugeth.RegisterModule[stateUpdateModule]()
}