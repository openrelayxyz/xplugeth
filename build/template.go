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
	"github.com/openrelayxyz/xplugeth/hooks/initialize"
	"github.com/openrelayxyz/xplugeth/hooks/modifyancients"
	"github.com/openrelayxyz/xplugeth/hooks/stateupdates"
	"github.com/openrelayxyz/xplugeth/hooks/triecommit"
	"github.com/openrelayxyz/xplugeth/types"
)

// init() registers the module with the plugin loader
func init() {
	xplugeth.RegisterModule[buildTemplateModule]("buildTemplateModule")
}

type buildTemplateModule struct {}

// GetAPIs fires as the node is registering its APIs. It should return a list of rpc.API objects so that plugins can register new RPC methods. This brings the plugin module into compliance with the apis.GetAPIs interface.
func (*buildTemplateModule) GetAPIs(stack *node.Node, backend types.Backend) []rpc.API {
	return nil
}

// NewHead fires when a new cannonical block is set. Note that during reorgs blocks can enter the canonical chain without this method being called on them. This brings the plugin module into compliance with the blockchain.NewHeadPlugin interface.
func (*buildTemplateModule) NewHead(block *gtypes.Block, hash common.Hash, logs []*gtypes.Log, totalDifficulty *big.Int) {
}

// Reorg fires when a chain reorg occurs. This brings the plugin module into compliance with the blockchain.ReorgPlugin interface.
func (*buildTemplateModule) Reorg(commonBlock common.Hash, oldChain []common.Hash, newChain []common.Hash) {
}

// NewSideBlock fires during the course of a reorg when a side block is set. This brings the plugin module into compliance with the blockchain.NewSideBlock interface.
func (*buildTemplateModule) NewSideBlock(block *gtypes.Block, hash common.Hash, logs []*gtypes.Log) {
}

// SetTrieFlushIntervalClone fires as the node evaluates the interval in which to flush the trie to disk. This brings the plugin module into compliance with the blockchain.SetTrieFlushIntervalClonePlugin interface.
func (*buildTemplateModule) SetTrieFlushIntervalClone(flushInterval time.Duration) time.Duration {
	var t time.Duration
	return t
}

// InitializeNode fires as the node starts. This brings the plugin module into compliance with the initialize.Initializer interface.
func (*buildTemplateModule) InitializeNode(stack *node.Node, backend types.Backend, cfg any) {
	log.Info("build template plugin initailized")
}

// Blockchain fires on start up after InitializeNode and acts as a generic trigger for various type of plugin functionality.
// This brings the plugin module into compliance with the initialize.Blockchain interface.
func (*buildTemplateModule) Blockchain() {
}

// Shutdown is defered until the node is shutting down. This brings the plugin module into compliance with the initialize.Shutdown interface.
func (*buildTemplateModule) Shutdown() {
}

// ModifyAncients as ancient write operations are commited. This brings the plugin module into compliance with the modifyancients.ModifyAncientsPlugin interface.
func (*buildTemplateModule) ModifyAncients(index uint64, header *gtypes.Header) {
}

// StateUpdate fires as state mutations are commited. This brings the plugin module into compliance with the stateupdates.StateUpdatePlugin interface.
func (*buildTemplateModule) StateUpdate(blockRoot common.Hash, parentRoot common.Hash, destructs map[common.Hash]struct{}, accounts map[common.Hash][]byte, storage map[common.Hash]map[common.Hash][]byte, codeUpdates map[common.Hash][]byte) {
}

// PreTrieCommit fires before all the children of a particular node are written to disk. This brings the plugin module into compliance with the triecommit.PreTrieCommit interface.
func (*buildTemplateModule) PreTrieCommit(node common.Hash) {
}

// PostTrieCommit fires after all the children of a particular node are written to disk. This brings the plugin module into compliance with the triecommit.PostTrieCommit interface.
func (*buildTemplateModule) PostTrieCommit(node common.Hash) {
}


// The following interface guards ensure that *buildTemplateModule implements the expected APIs correctly.
// Interface guards are not strictly required, but simply registering a module with the plugin loader does
// not guarantee interface compliance, and if an interface is not implemented correctly the hooks will
// quietly fail to be invoked.
//
// We recommend using interface guards to do compile-time checks that a module implements the interfaces
// it is intended to implement.
var (
	_ apis.GetAPIs = (*buildTemplateModule)(nil)
	_ blockchain.NewHeadPlugin = (*buildTemplateModule)(nil)
	_ blockchain.NewSideBlockPlugin = (*buildTemplateModule)(nil)
	_ blockchain.ReorgPlugin = (*buildTemplateModule)(nil)
	_ blockchain.SetTrieFlushIntervalClonePlugin = (*buildTemplateModule)(nil)
	_ initialize.Initializer = (*buildTemplateModule)(nil)
	_ initialize.Blockchain = (*buildTemplateModule)(nil)
	_ initialize.Shutdown = (*buildTemplateModule)(nil)
	_ modifyancients.ModifyAncientsPlugin = (*buildTemplateModule)(nil)
	_ stateupdates.StateUpdatePlugin = (*buildTemplateModule)(nil)
	_ triecommit.PreTrieCommit = (*buildTemplateModule)(nil)
	_ triecommit.PostTrieCommit = (*buildTemplateModule)(nil)
)
