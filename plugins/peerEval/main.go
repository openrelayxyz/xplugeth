package peereval

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	gtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/openrelayxyz/xplugeth"
	"github.com/openrelayxyz/xplugeth/hooks/apis"
	"github.com/openrelayxyz/xplugeth/hooks/fetcher"
	"github.com/openrelayxyz/xplugeth/hooks/initialize"
	"github.com/openrelayxyz/xplugeth/types"
)

type PeerMetrics struct {
	ID                    string
	BlocksContributed     int
	ContributionIntervals map[int]bool
	LastConnected         time.Time
	LastDisconnected      time.Time
	IsConnected           bool
	ConnectedTime         time.Duration
}

var (
	client         *rpc.Client
	gatheringCount int
	called         bool
	activePeerData []map[string]interface{}

	intervalDuration = 10
	currentInterval  = 0
	monitoringPeriod = 100
	blockCount       = 0
	pollingInterval  = time.Minute
	averageBlockTime = 15.0
)

type peerEvalPlugin struct {
	peerMetricsMap map[string]*PeerMetrics
	mutex          sync.Mutex
}

func init() {
	xplugeth.RegisterModule[peerEvalPlugin]("peerEvalPlugin")
}

func (p *peerEvalPlugin) InitializeNode(s *node.Node, _ types.Backend) {
	client = s.Attach()
	p.peerMetricsMap = make(map[string]*PeerMetrics)
	p.StartPeerMonitoring()
}

func getPeers() ([]string, error) {
	var rawPeerData []map[string]interface{}
	err := client.Call(&rawPeerData, "admin_peers")
	if err != nil {
		log.Error("error calling admin_peers, peerEval plugin", "err", err)
		return nil, err
	}

	peers := []string{}

	for _, item := range rawPeerData {
		for k, v := range item {
			if k == "id" {
				peers = append(peers, v.(string))
			}
		}
	}

	return peers, nil
}

func (p *peerEvalPlugin) PeerEval(id string, headers []*gtypes.Header) {
	gatheringCount++
	blockCount += len(headers)
	log.Error(fmt.Sprintf("blockcount %v", blockCount))

	currentInterval = blockCount / intervalDuration

	if _, exists := p.peerMetricsMap[id]; !exists {
		p.peerMetricsMap[id] = &PeerMetrics{
			ID:                    id,
			ContributionIntervals: make(map[int]bool),
			IsConnected:           true,
			LastConnected:         time.Now(),
		}
	}

	peerMetric := p.peerMetricsMap[id]
	peerMetric.BlocksContributed += len(headers)
	peerMetric.ContributionIntervals[currentInterval] = true

	if blockCount >= monitoringPeriod {
		blockCount = 0
		currentInterval = 0
		p.evaluatePeers()
		p.resetMetrics()
	}

	// log.Error(fmt.Sprintf("Gathering peer data, count %v/100", gatheringCount))

	// t := time.Now().Format("20060102_150405")

	// blockNumbers := []string{}
	// for _, header := range headers {
	// 	blockNumbers = append(blockNumbers, header.Number.String())
	// }
	// evalData := map[string]interface{}{
	// 	"id":     id,
	// 	"time":   t,
	// 	"blocks": blockNumbers,
	// }
	// activePeerData = append(activePeerData, evalData)
	// if gatheringCount >= 100 {
	// 	gatheringCount = 0
	// 	if called {
	// 		called = false
	// 		returnPeerData()
	// 	}
	// 	activePeerData = activePeerData[:0]
	// }
}

func (p *peerEvalPlugin) StartPeerMonitoring() {
	ticker := time.NewTicker(pollingInterval)
	go func() {
		for range ticker.C {
			p.updatePeerConnections()
		}
	}()
}

func (p *peerEvalPlugin) updatePeerConnections() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	peerIDs, err := getPeers()
	if err != nil {
		log.Error("Failed to get peers", "err", err)
		return
	}

	currentPeers := make(map[string]bool)
	for _, id := range peerIDs {
		currentPeers[id] = true

		peerMetric, exists := p.peerMetricsMap[id]
		if !exists {
			p.peerMetricsMap[id] = &PeerMetrics{
				ID:                    id,
				ContributionIntervals: make(map[int]bool),
				IsConnected:           true,
				LastConnected:         time.Now(),
			}
		} else if !peerMetric.IsConnected {

			peerMetric.IsConnected = true
			peerMetric.LastConnected = time.Now()
		}
	}
	for id, peerMetric := range p.peerMetricsMap {
		if !currentPeers[id] && peerMetric.IsConnected {
			peerMetric.IsConnected = false
			peerMetric.LastDisconnected = time.Now()
			peerMetric.ConnectedTime += peerMetric.LastDisconnected.Sub(peerMetric.LastConnected)
		}
	}
}

func (p *peerEvalPlugin) evaluatePeers() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for _, peerMetric := range p.peerMetricsMap {
		consistencyScore := len(peerMetric.ContributionIntervals)

		if peerMetric.IsConnected {
			peerMetric.ConnectedTime += time.Since(peerMetric.LastConnected)
			peerMetric.LastConnected = time.Now()
		}

		totalElapsedTime := time.Since(peerMetric.LastConnected) + peerMetric.ConnectedTime
		uptimePercentage := (peerMetric.ConnectedTime.Seconds() / totalElapsedTime.Seconds()) * 100

		log.Error(fmt.Sprintf("Peer Metrics\nID: %s\nBlocksContributed: %d\nConsistencyScore: %d\nUptimePercentage: %.2f",
			peerMetric.ID,
			peerMetric.BlocksContributed,
			consistencyScore,
			uptimePercentage,
		))
	}
}

func (p *peerEvalPlugin) resetMetrics() {
	for _, peerMetric := range p.peerMetricsMap {
		peerMetric.BlocksContributed = 0
		peerMetric.ContributionIntervals = make(map[int]bool)
		// peerMetric.ConnectedTime = 0
		// peerMetric.LastConnected = time.Now()
		// peerMetric.IsConnected = true
	}
}

func returnPeerData() {
	log.Error("gathering peer data for return")
	peerSlice, err := getPeers()
	if err != nil {
		log.Error("error obtaining peer slice", "err", err)

	}
	data := make(map[string]interface{})
	data["peers"] = peerSlice
	data["active"] = activePeerData

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Error("error marshaling JSON", "err", err)
	}

	file, err := os.Create(fmt.Sprintf("peer-data-%v.json", time.Now().Format("20060102_150405")))
	if err != nil {
		log.Error("error creating file", "err", err)
	}
	defer file.Close()

	_, err = file.Write(jsonData)
	if err != nil {
		log.Error("error writing to file return PeerData", "err", err)
	}
}

type peerEvalAPI struct{}

func (p *peerEvalAPI) GetPeerData() string {
	called = true
	return "signal set"
}

func (p *peerEvalPlugin) GetAPIs(*node.Node, types.Backend) []rpc.API {
	log.Info("Registering peer eval plugin APIs")
	return []rpc.API{
		{
			Namespace: "plugeth",
			Service:   &peerEvalAPI{},
		},
	}
}

var (
	_ apis.GetAPIs           = (*peerEvalPlugin)(nil)
	_ initialize.Initializer = (*peerEvalPlugin)(nil)
	_ fetcher.PeerEvalPlugin = (*peerEvalPlugin)(nil)
)
