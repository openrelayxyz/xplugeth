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

	// for each peer we need to know its id, and a contribution ratio = blocks contributed / duration 
	// when we evalutate at a set interval we will cast off the bottom 10% in terms of contributions. 
	
	// {id, blocks contributed, total time connected}

	// we will evaluate the contribution ratio each time we evaluate peers


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
	maxPeers int

	// Jesse, bellow is the flag implementation
	flags = *flag.NewFlagSet("peereval-plugin", flag.ContinueOnError)
	maxPeerCount = flags.Int("peereval.max.peers", 0, "max peer value for peer eval plugin")
	// lets make polling interval and connection time coefficient a flag value
)

type peerEvalPlugin struct {
	peerMetricsMap map[string]*PeerMetrics
	mutex          sync.Mutex
}

func init() {
	xplugeth.RegisterModule[peerEvalPlugin]("peerEvalPlugin")
}

func (p *peerEvalPlugin) InitializeNode(s *node.Node, _ types.Backend) {

	//Jesse flag work below
	if *maxPeerCount == 0 {
		log.Warn("max peer count flag not set, peer eval plugin, setting to default of 20")
		maxPeers = 20
	}

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
	// ultimate job of peer eval is to augment the block count for the peer
	
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

	// cast off any non connected peers from the map
	// evaluate the peer: create the contribution ratio
	// cast off the bottom 10% performing peers
	
	// two conditions have to be met before we cast off peers:
	// ONE: our peer map has to be above 90% of what the max peer count, use maxPeers to evaluate this

	// TWO: peer needs to have been connected for greater than the connection time coefficient 
	// connection time coefficient is a factor that we multiply the polling interval by
	// lets set it to 5 minutes as a default

	// going to need to make a function that disconnects peers, it is in the admin name space *I think



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


type peerEvalAPI struct{}

func (p *peerEvalAPI) GetPeerData() string {
	called = true
	return "signal set"
}

func (p *peerEvalAPI) TestPeerEval() string {
	return "calling from peer eval"
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
