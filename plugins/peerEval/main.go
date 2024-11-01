package peereval

import (
	"github.com/ethereum/go-ethereum/common"
	gtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/openrelayxyz/xplugeth"
	"github.com/openrelayxyz/xplugeth/hooks/blockchain"
	"github.com/openrelayxyz/xplugeth/hooks/initialize"
	"github.com/openrelayxyz/xplugeth/types"
)

var (
	stack  node.Node
	client *rpc.Client
)

type peerEvalPlugin struct {
}

func init() {
	xplugeth.RegisterModule[peerEvalPlugin]("peerEvalModule")
}

func (p *peerEvalPlugin) InitializeNode(s *node.Node, b types.Backend) {
	stack = *s
	client = stack.Attach()
}

func (p *peerEvalPlugin) getPeers() ([]string, error) {
	var peers []map[string]interface{}
	err := client.Call(&peers, "admin_peers")
	if err != nil {
		log.Error("error calling admin_peers, peerEval plugin", "err", err)
		return nil, err
	}

	var peerIds []string
	for _, peer := range peers {
		if id, ok := peer["id"].(string); ok {
			peerIds = append(peerIds, id)
		}
	}
	return peerIds, nil
}

func (p *peerEvalPlugin) PeerEval(id string, headers []*gtypes.Header, hashes []common.Hash) {
	peers, _ := p.getPeers()
	for _, peerID := range peers {
		log.Info("Peer ID", "id", peerID)
	}
}

var (
	_ initialize.Initializer    = (*peerEvalPlugin)(nil)
	_ blockchain.PeerEvalPlugin = (*peerEvalPlugin)(nil)
)
