package peereval

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	gtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/openrelayxyz/xplugeth"
	"github.com/openrelayxyz/xplugeth/hooks/apis"
	"github.com/openrelayxyz/xplugeth/hooks/blockchain"
	"github.com/openrelayxyz/xplugeth/hooks/initialize"
	"github.com/openrelayxyz/xplugeth/types"
)

var (
	stack    node.Node
	client   *rpc.Client
	peerData []map[string]interface{}
	count    int
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

func (p *peerEvalPlugin) getPeers() ([]map[string]string, error) {
	var peers []map[string]interface{}
	err := client.Call(&peers, "admin_peers")
	if err != nil {
		log.Error("error calling admin_peers, peerEval plugin", "err", err)
		return nil, err
	}

	peerMapList := make([]map[string]string, 0)
	for _, peer := range peers {
		if id, ok := peer["id"].(string); ok {
			if enode, ok := peer["enode"].(string); ok {
				peerMap := map[string]string{
					id: enode,
				}
				peerMapList = append(peerMapList, peerMap)
			}
		}
	}
	return peerMapList, nil

}

func (*peerEvalPlugin) NewHead(block *gtypes.Block, hash common.Hash, logs []*gtypes.Log, td *big.Int) {
	log.Error("these are the fields in question", "id", block.ReceivedFrom, "time", block.ReceivedAt)
}

type peerEvalAPI struct {}

func (*peerEvalPlugin) GetAPIs(*node.Node, types.Backend) []rpc.API {
	log.Info("Registering peer eval APIs")
	return []rpc.API{
		{
			Namespace: "plugeth",
			Service:   &peerEvalAPI{},
		},
	}
}

func (*peerEvalAPI) GetCount() int {
	return count
}

var (
	_ initialize.Initializer    = (*peerEvalPlugin)(nil)
	_ blockchain.NewHeadPlugin = (*peerEvalPlugin)(nil)
	// _ blockchain.PeerEvalPlugin = (*peerEvalPlugin)(nil)
	// _ initialize.Blockchain     = (*peerEvalPlugin)(nil)
	_ apis.GetAPIs				= (*peerEvalPlugin)(nil)
)
