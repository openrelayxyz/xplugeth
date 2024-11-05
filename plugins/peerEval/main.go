package peereval

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

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

func (p *peerEvalPlugin) Blockchain() {
	peers, err := p.getPeers()
	if err != nil {
		log.Error("failed to get peers", "err", err)
		return
	}
	for _, peer := range peers {
		for id, enode := range peer {
			log.Info("Peer found", "id", id, "enode", enode)
		}
	}
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

func (p *peerEvalPlugin) PeerEval(id string, headers []*gtypes.Header, hashes []common.Hash) {
	timeNow := time.Now().Format("15:04:05")

	blockNumbers := make([]uint64, 0)
	for _, header := range headers {
		blockNumbers = append(blockNumbers, header.Number.Uint64())
	}

	evalData := map[string]interface{}{
		timeNow: map[string]interface{}{
			"id":     id,
			"blocks": blockNumbers,
		},
	}
	peerData = append(peerData, evalData)

	count++

	if count >= 20 {
		jsonData, _ := json.MarshalIndent(peerData, "", "  ")
		filename := fmt.Sprintf("peer_eval_%s.json", time.Now().Format("20060102_150405"))
		if err := os.WriteFile(filename, jsonData, 0644); err != nil {
			log.Error("failed to write data to file", "err", err)
			return
		}
	}

}

var (
	_ initialize.Initializer    = (*peerEvalPlugin)(nil)
	_ blockchain.PeerEvalPlugin = (*peerEvalPlugin)(nil)
	_ initialize.Blockchain     = (*peerEvalPlugin)(nil)
)
