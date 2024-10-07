package triecommit

import (
	"github.com/openrelayxyz/xplugeth"
	
	"github.com/ethereum/go-ethereum/common"
)

type PreTrieCommit interface {
	PreTrieCommit(node common.Hash)
}

type PostTrieCommit interface {
	PostTrieCommit(node common.Hash)
}

func init() {
	xplugeth.RegisterHook[PreTrieCommit]()
	xplugeth.RegisterHook[PostTrieCommit]()
}