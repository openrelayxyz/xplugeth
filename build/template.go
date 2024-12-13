package build

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	gtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/openrelayxyz/xplugeth"
	"github.com/openrelayxyz/xplugeth/hooks/apis"
	"github.com/openrelayxyz/xplugeth/hooks/blockchain"
	"github.com/openrelayxyz/xplugeth/hooks/fetcher"
	"github.com/openrelayxyz/xplugeth/hooks/initialize"
	"github.com/openrelayxyz/xplugeth/hooks/modifyancients"
	"github.com/openrelayxyz/xplugeth/hooks/stateupdates"
	"github.com/openrelayxyz/xplugeth/hooks/triecommit"
	"github.com/openrelayxyz/xplugeth/types"
)

func init() {
	xplugeth.RegisterModule[buildTemplateModule]("buildTemplateModule")
}

type buildTemplateModule struct {}

func (*buildTemplateModule) GetAPIs(*node.Node, types.Backend) []rpc.API {
	return nil
}

func (*buildTemplateModule) NewHead(*gtypes.Block, common.Hash, []*gtypes.Log, *big.Int) {
}

func (*buildTemplateModule) Reorg(common.Hash, []common.Hash, []common.Hash) {
}

func (*buildTemplateModule) NewSideBlock(*gtypes.Block, common.Hash, []*gtypes.Log) {
}

func (*buildTemplateModule) SetTrieFlushIntervalClone(time.Duration) time.Duration {
	var t time.Duration
	return t
}

func (*buildTemplateModule) InitializeNode(*node.Node, types.Backend) {
	log.Info("build template plugin initailized")
}

func (*buildTemplateModule) Blockchain() {
}

func (*buildTemplateModule) Shutdown() {
}

func (*buildTemplateModule) ModifyAncients(uint64, *gtypes.Header) {
}

func (*buildTemplateModule) StateUpdate(common.Hash, common.Hash, map[common.Hash]struct{}, map[common.Hash][]byte, map[common.Hash]map[common.Hash][]byte, map[common.Hash][]byte) {
}

func (*buildTemplateModule) PreTrieCommit(common.Hash) {
}

func (*buildTemplateModule) PostTrieCommit(common.Hash) {
}



var (
	_ apis.GetAPIs = (*buildTemplateModule)(nil)
	_ blockchain.NewHeadPlugin = (*buildTemplateModule)(nil)
	_ blockchain.NewSideBlockPlugin = (*buildTemplateModule)(nil)
	_ blockchain.ReorgPlugin = (*buildTemplateModule)(nil)
	_ blockchain.SetTrieFlushIntervalClonePlugin = (*buildTemplateModule)(nil)
	_ fetcher.PeerEvalPlugin = (*buildTemplateModule)(nil)
	_ initialize.Initializer = (*buildTemplateModule)(nil)
	_ initialize.Blockchain = (*buildTemplateModule)(nil)
	_ initialize.Shutdown = (*buildTemplateModule)(nil)
	_ modifyancients.ModifyAncientsPlugin = (*buildTemplateModule)(nil)
	_ stateupdates.StateUpdatePlugin = (*buildTemplateModule)(nil)
	_ triecommit.PreTrieCommit = (*buildTemplateModule)(nil)
	_ triecommit.PostTrieCommit = (*buildTemplateModule)(nil)
)