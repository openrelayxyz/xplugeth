package peereval

import (
	"encoding/json"
	"flag"
	"sort"
	"sync"
	"time"

	"github.com/Shopify/sarama"

	gtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/openrelayxyz/xplugeth"
	"github.com/openrelayxyz/xplugeth/hooks/fetcher"
	"github.com/openrelayxyz/xplugeth/hooks/initialize"
	"github.com/openrelayxyz/xplugeth/types"
	xp_utils "github.com/openrelayxyz/xplugeth/utils/plugins"
	"github.com/openrelayxyz/xplugeth/plugins/peermanager"
)

type PeerMetrics struct {
	ID                string
	Enode             string
	BlocksContributed int
	LastConnected     time.Time
	LastDisconnected  time.Time
	IsConnected       bool
	ConnectedTime     time.Duration
	IsInbound 		  bool 
}

var (
	blockCount int
	client  *rpc.Client 
	chainid int64
	broker string
	topic string
	flags                     = *flag.NewFlagSet("peereval-plugin", flag.ContinueOnError)
	maxPeerCount              = flags.Int("peereval.max.peers", 10, "max peer value for peer eval plugin")
	pollingInterval           = flags.Duration("peereval.polling.interval", time.Minute, "polling interval for peer monitoring")
	connectionTimeCoefficient = flags.Duration("peereval.connection.time.coefficient", 5*time.Minute, "minimum connection time before evaluating peers")
)

type peerEvalModule struct {
	peerMetricsMap map[string]*PeerMetrics
	mutex          sync.RWMutex
	peerRatios     map[string]float64
	producer       sarama.AsyncProducer
}

func init() {
	xplugeth.RegisterModule[peerEvalModule]("peerEvalModule")
	xplugeth.RegisterFlags(flags)
}

func (p *peerEvalModule) InitializeNode(s *node.Node, b types.Backend, c any) {
	client =  s.Attach()
		
	p.peerMetricsMap = make(map[string]*PeerMetrics)
	p.StartPeerMonitoring()
	p.cleanUpPeerMap()
		
	log.Info("Initialized node, peer eval plugin")
}

func (p *peerEvalModule) Blockchain () {
	if peermanager.SharedBroker != nil {
		broker = *peermanager.SharedBroker
		topic = *peermanager.SharedTopic
		producer, err := xp_utils.CreateProducer(broker, topic)
		if err != nil {
			log.Error("failed to create Kafka producer", "err", err)
			return
		}
		p.producer = producer
		go p.streamHealthyPeers()
	} else {
		log.Warn("peer evaluation sequence not available")
	}
}



type peerInfo struct {
	ID string
	Enode string
	Inbound bool
}

func getPeers() ([]peerInfo, error) {
	var rawPeerData []map[string]interface{}
	err := client.Call(&rawPeerData, "admin_peers")
	if err != nil {
		log.Error("error calling admin_peers, peerEval plugin", "err", err)
		return nil, err
	}

	peers := []peerInfo{}

	for _, item := range rawPeerData {
		var id, enode string
		var inbound bool
		for k, v := range item {
			if k == "id" {
				id = v.(string)
			}
			if k == "enode" {
				enode = v.(string)
			}
			if k == "network" {
				network := v.(map[string]interface{})
				inbound = network["inbound"].(bool)
			}
		}

		if id != ""{
			peers = append(peers, peerInfo{
				ID: id,
				Enode: enode,
				Inbound: inbound,
			})
		}
	}

	return peers, nil
}

func (p *peerEvalModule) PeerEval(id string, headers []*gtypes.Header) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

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

func (p *peerEvalModule) StartPeerMonitoring() {
	ticker := time.NewTicker(*pollingInterval)
	go func() {
		
		for range ticker.C {
			p.updatePeerConnections()
		}
	}()
}

func (p *peerEvalModule) cleanUpPeerMap() {
	ticker := time.NewTicker(*pollingInterval)
	go func() {
		for range ticker.C {
			p.mutex.Lock()
			for id, peer := range p.peerMetricsMap {
				connectedDuration := peer.ConnectedTime + time.Since(peer.LastConnected)
				if !peer.IsConnected && connectedDuration >= *connectionTimeCoefficient {
					delete(p.peerMetricsMap, id)
				}
			}
			p.mutex.Unlock()
		}
	}()
}

func (p *peerEvalModule) updatePeerConnections() {
	peers, err := getPeers()
	if err != nil {
		log.Error("Failed to get peers", "err", err)
		return
	}

	currentPeers := make(map[string]bool)

	p.mutex.Lock()
	for _, peer := range peers {
		currentPeers[peer.ID] = true

		peerMetric, exists := p.peerMetricsMap[peer.ID]
		if !exists {
			p.peerMetricsMap[peer.ID] = &PeerMetrics{
				ID:            peer.ID,
				Enode: 		   peer.Enode,
				IsConnected:   true,
				LastConnected: time.Now(),
				IsInbound:     peer.Inbound,
			}
		} else if !peerMetric.IsConnected {
			peerMetric.IsConnected = true
			peerMetric.LastConnected = time.Now()
			peerMetric.IsInbound = peer.Inbound
			peerMetric.Enode = peer.Enode
		}
	}
	
	for id, peerMetric := range p.peerMetricsMap {
		if !currentPeers[id] && peerMetric.IsConnected {
			peerMetric.IsConnected = false
			peerMetric.LastDisconnected = time.Now()
			peerMetric.ConnectedTime += peerMetric.LastDisconnected.Sub(peerMetric.LastConnected)
		}
	}
	p.mutex.Unlock()

	p.mutex.Lock()
	p.peerRatios = make(map[string]float64)
		for id, peer := range p.peerMetricsMap {
			if peer.IsConnected {
				connectedDuration := peer.ConnectedTime + time.Since(peer.LastConnected)
				if connectedDuration >= *connectionTimeCoefficient {
					p.peerRatios[id] = float64(peer.BlocksContributed) / connectedDuration.Seconds()
			}
		}
	}
	p.mutex.Unlock()

	p.mutex.RLock()
	peerCount := len(p.peerMetricsMap)
	outboundCount := 0

	for _, peer := range p.peerMetricsMap{
		if !peer.IsInbound{
			outboundCount++
		}
	}
	p.mutex.RUnlock()

	if peerCount > int(float64(*maxPeerCount)*0.9) {
		p.prunePeers(true)
	} else if float64(outboundCount) >= float64(peerCount) * 0.9 {
			p.prunePeers(false)
	}
}

func (p *peerEvalModule) getHealthyPeers() []string{
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	healthy := []string{""}
	for _, metrics := range p.peerMetricsMap {
		if metrics.BlocksContributed > 0 && metrics.Enode != "" {
			healthy = append(healthy, metrics.Enode)
		}
	}
	return healthy
}

func (p *peerEvalModule) streamHealthyPeers() {
	ticker := time.NewTicker(4 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		peers := p.getHealthyPeers()
		if len(peers) == 0 {
			continue
		}
		data, err := json.Marshal(peers)
		if err != nil {
			log.Error("failed to marshal healthy peers", "err", err)
			continue
		}
		msg := &sarama.ProducerMessage{
			Topic : topic,
			Value : sarama.ByteEncoder(data),
		}
		log.Info("sending getHealthyPeers, peerevaluator", "length", len(peers))
		p.producer.Input() <-msg 
	}
}

func (p *peerEvalModule) prunePeers(pruneAll bool){
	if len(p.peerRatios) == 0 {
		return
	}
	var peersToDrop []string

	p.mutex.RLock()
	for id, peer := range p.peerMetricsMap {
		if pruneAll || !peer.IsInbound {
			peersToDrop = append(peersToDrop, id)
		}
	}
	p.mutex.RUnlock()

	dropCount := int(0.1 * float64(len(peersToDrop)))
	if dropCount == 0 {
		return
	}
	sort.Slice(peersToDrop, func(i, j int) bool {
		return p.peerRatios[peersToDrop[i]] < p.peerRatios[peersToDrop[j]]
	})
	p.removePeers(peersToDrop[:dropCount])
}

func (p *peerEvalModule) removePeers(peers []string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	for _, id := range peers {
		metrics, ok := p.peerMetricsMap[id]
		if !ok || metrics.Enode == "" {
			continue
		}
		var result bool
		if err := client.Call(&result, "admin_removePeer", metrics.Enode ); err != nil {
			log.Error("Failed to remove peer", "enode", metrics.Enode, "err", err)
		} else {
			log.Info("Removed peer", "enode", metrics.Enode)
			delete(p.peerMetricsMap, id)
		}
	}
}

var (
	_ initialize.Initializer = (*peerEvalModule)(nil)
	_ fetcher.PeerEvalPlugin = (*peerEvalModule)(nil)
)
