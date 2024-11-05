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
	xplugeth.RegisterHook[PreTrieCommit](
		xplugeth.Patchset{
			Remote: "github.com/openrelayxyz/xplugeth-patches",
			Ref: "hooks_triecommit_foundation_v1.14.11_1",
			Tests: []xplugeth.Test{
				{
					Package: "./core",
					TestNames: []string{
						"TestHashDBInjections",
					},
				},
			},
		},
	)
	xplugeth.RegisterHook[PostTrieCommit]()
}