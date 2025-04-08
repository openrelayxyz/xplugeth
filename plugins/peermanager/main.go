package peermanager

import (
	"flag"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/Shopify/sarama"

	gtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/openrelayxyz/xplugeth"
	"github.com/openrelayxyz/xplugeth/hooks/apis"
	"github.com/openrelayxyz/xplugeth/hooks/fetcher"
	"github.com/openrelayxyz/xplugeth/hooks/initialize"
	"github.com/openrelayxyz/xplugeth/types"
	"github.com/openrelayxyz/xplugeth/utils"
)

type peerManagerConfig struct {
	BrokerURL string `yaml:"broker.url"`
	PeerTopic string `yaml:"peer.topic"`
}

type PeerMetrics struct {
	ID                string
	BlocksContributed int
	LastConnected     time.Time
	LastDisconnected  time.Time
	IsConnected       bool
	ConnectedTime     time.Duration
}

var (
	blockCount = 0
	maxPeers   int

	sessionPeerService *PeerManager
	chainid            int64
	brokers            []string
	config             *sarama.Config
	nodes              = make(chan string, 5)
	exit               = make(chan struct{}, 1)
	cfg                *peerManagerConfig
	peerBroker         string
	peerTopic          string

	flags                     = *flag.NewFlagSet("peereval-plugin", flag.ContinueOnError)
	maxPeerCount              = flags.Int("peereval.max.peers", 0, "max peer value for peer eval plugin")
	pollingInterval           = flags.Duration("peereval.polling.interval", time.Minute, "polling interval for peer monitoring")
	connectionTimeCoefficient = flags.Duration("peereval.connection.time.coefficient", 5*time.Minute, "minimum connection time before evaluating peers")
)

type peerManagerModule struct {
	peerMetricsMap map[string]*PeerMetrics
	mutex          sync.Mutex
}

func init() {
	xplugeth.RegisterModule[peerManagerModule]("peerManagerModule")
}

func (p *peerManagerModule) InitializeNode(s *node.Node, b types.Backend) {

	sessionPeerService = &PeerManager{
		client: s.Attach(),
	}

	var ok bool

	chainid, ok = utils.GetChainID()
	if !ok {
		panic(fmt.Sprintf("could not resolve chain id from xplugeth utils, peermanager"))
	}

	cfg, ok = xplugeth.GetConfig[peerManagerConfig]("peermanager")
	if !ok {
		cfg = &peerManagerConfig{}
		log.Warn("did not acqire config, example plugin, all values set to default")
	}
	peerBroker = cfg.BrokerURL
	peerTopic = cfg.PeerTopic

	if *maxPeerCount == 0 {
		log.Warn("max peer count flag not set, peer eval plugin, setting to a default of 20")
		maxPeers = 20
	}
	p.peerMetricsMap = make(map[string]*PeerMetrics)
	p.StartPeerMonitoring()
	p.cleanUpPeerMap()

	log.Info("Initialized node, peer manager plugin")
}

func (*peerManagerModule) Blockchain() {
	if sessionPeerService == nil {
		panic(fmt.Sprintf("peer manager is nil, peer manager plugin"))
	}
	go peeringSequence()
}

func peeringSequence() {

	selfNode, err := sessionPeerService.getEnode()
	if err != nil {
		log.Error("error calling getEnode from sessionService, peer manager plugin", "err", err)
	}

	producer, err := createProducer(peerBroker, peerTopic)
	if err != nil {
		log.Error("failed to acquire kafka producer, peer manager plugin", "err", err)
		return
	}

	consumer, err := createConsumer(peerBroker, peerTopic)
	if err != nil {
		log.Error("failed to acquire kafka consumer, peer manager plugin", "err", err)
		return
	}

	msg := &sarama.ProducerMessage{
		Topic: peerTopic,
		Value: sarama.StringEncoder(selfNode),
	}

	producer.Input() <- msg

	go func() {
		for message := range consumer.Messages() {
			nodes <- string(message.Value)
		}
	}()

	for message := range nodes {
		if message == selfNode {
			continue
		} else {
			sessionPeerService.attachPeers(message)
		}
	}
}

func getPeers() ([]string, error) {
	var rawPeerData []map[string]interface{}
	err := sessionPeerService.client.Call(&rawPeerData, "admin_peers")
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

func (p *peerManagerModule) PeerEval(id string, headers []*gtypes.Header) {
	blockCount += len(headers)

	if _, exists := p.peerMetricsMap[id]; !exists {
		p.peerMetricsMap[id] = &PeerMetrics{
			ID:            id,
			IsConnected:   true,
			LastConnected: time.Now(),
		}
	}

	peerMetric := p.peerMetricsMap[id]
	peerMetric.BlocksContributed += len(headers)
}

func (p *peerManagerModule) StartPeerMonitoring() {
	ticker := time.NewTicker(*pollingInterval)
	go func() {
		for range ticker.C {
			p.updatePeerConnections()
		}
	}()
}

func (p *peerManagerModule) cleanUpPeerMap() {
	ticker := time.NewTicker(*pollingInterval)
	go func() {
		for range ticker.C {
			for id, peer := range p.peerMetricsMap {
				connectedDuration := peer.ConnectedTime + time.Since(peer.LastConnected)
				if !peer.IsConnected && connectedDuration >= *connectionTimeCoefficient {
					delete(p.peerMetricsMap, id)
				}
			}
		}
	}()

}

func (p *peerManagerModule) updatePeerConnections() {
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
				ID:            id,
				IsConnected:   true,
				LastConnected: time.Now(),
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

	if len(p.peerMetricsMap) < int(float64(maxPeers)*0.9) {
		return
	}

	peerRatios := make(map[string]float64)
	var peersToDrop []string

	for id, peer := range p.peerMetricsMap {
		if peer.IsConnected {
			connectedDuration := peer.ConnectedTime + time.Since(peer.LastConnected)
			if connectedDuration >= *connectionTimeCoefficient {
				peerRatios[id] = float64(peer.BlocksContributed) / connectedDuration.Seconds()
			}
		}
	}

	if len(peerRatios) == 0 {
		return
	}

	dropCount := int(float64(len(peerRatios)) * 0.1)
	if dropCount > 0 {
		peers := make([]string, 0, len(peerRatios))
		for id := range peerRatios {
			peers = append(peers, id)
		}
		sort.Slice(peers, func(i, j int) bool {
			return peerRatios[peers[i]] < peerRatios[peers[j]]
		})
		peersToDrop = peers[:dropCount]
		p.removePeers(peersToDrop)

	}

}

func (p *peerManagerModule) removePeers(peers []string) {
	for _, id := range peers {
		var result bool
		if err := sessionPeerService.client.Call(&result, "admin_removePeer", fmt.Sprintf("enode://%s", id)); err != nil {
			log.Error("Failed to remove peer", "id", id, "err", err)
		} else {
			log.Info("Removed peer", "id", id)
			delete(p.peerMetricsMap, id)
		}
	}
}

type peerManagerAPI struct{}

func (p *peerManagerModule) GetAPIs(*node.Node, types.Backend) []rpc.API {
	log.Info("Registering peer eval plugin APIs")
	return []rpc.API{
		{
			Namespace: "plugeth",
			Service:   &peerManagerAPI{},
		},
	}
}

var (
	_ apis.GetAPIs           = (*peerManagerModule)(nil)
	_ initialize.Blockchain  = (*peerManagerModule)(nil)
	_ initialize.Initializer = (*peerManagerModule)(nil)
	_ fetcher.PeerEvalPlugin = (*peerManagerModule)(nil)
)
