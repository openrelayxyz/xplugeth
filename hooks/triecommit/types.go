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
	xplugeth.RegisterHook[PreTrieCommit](
		xplugeth.Patchset{
			Remote: "github.com/openrelayxyz/xplugeth-patches",
			Ref: "hooks_foundation_triecommit_v1.14.11_0",
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