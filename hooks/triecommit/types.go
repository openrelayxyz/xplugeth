package triecommit

import (
	"github.com/openrelayxyz/xplugeth"
	
	"github.com/ethereum/go-ethereum/common"
)

type PreTrieCommit interface {
	PreTrieCommit(common.Hash)
}

type PostTrieCommit interface {
	PostTrieCommit(common.Hash)
}

func init() {
	xplugeth.RegisterHook[PreTrieCommit](triecommitPatchsets...)
	xplugeth.RegisterHook[PostTrieCommit]()
}