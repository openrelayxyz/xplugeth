package peereval

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	gtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/openrelayxyz/xplugeth"
	"github.com/openrelayxyz/xplugeth/hooks/blockchain"
	"github.com/openrelayxyz/xplugeth/hooks/initialize"
	"github.com/openrelayxyz/xplugeth/types"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
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
	xplugeth.RegisterModule[peerEvalPlugin]("peerEvalPlugin")
}

func (p *peerEvalPlugin) InitializeNode(s *node.Node, b types.Backend) {
	stack = *s
	client = stack.Attach()

	go func() {
		for {
			cpuPercent, err := cpu.Percent(0, false)
			if err != nil {
				log.Error("Failed to get CPU stats", "err", err)
				continue
			}
			v, err := mem.VirtualMemory()
			if err != nil {
				log.Error("Failed to get memory stats", "err", err)
				continue
			}

			log.Info("System Stats",
				"cpu_usage", fmt.Sprintf("%.2f%%", cpuPercent[0]),
				"used memory", fmt.Sprintf("%.2f MB", float64(v.Used)/1024/1024),
				"total memory", fmt.Sprintf("%.2f MB", float64(v.Total)/1024/1024),
				"free memory", fmt.Sprintf("%.2f MB", float64(v.Free)/1024/1024),
				"memory usage", fmt.Sprintf("%.2f%%", v.UsedPercent),
			)
			time.Sleep(15 * time.Second)
		}
	}()
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

func (p *peerEvalPlugin) PeerEval(id string, headers []*gtypes.Header) {
	log.Error("inside of peer eval")
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
	log.Error("hit", "no", count)
	if count >= 10 {
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
