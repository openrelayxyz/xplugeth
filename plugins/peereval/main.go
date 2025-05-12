package peereval

import (
	"encoding/json"
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
	"github.com/openrelayxyz/xplugeth/hooks/fetcher"
	"github.com/openrelayxyz/xplugeth/hooks/initialize"
	"github.com/openrelayxyz/xplugeth/types"
	"github.com/openrelayxyz/xplugeth/utils"
)

type peerEvalConfig struct {
	BrokerURL string `yaml:"broker.url"`
	Topic string `yaml:"topic"`
}

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

type HealthyPeers interface {
	GetHealthyPeers() []string
}

var (
	blockCount = 0
	client  *rpc.Client 
	chainid            int64
	cfg  *peerEvalConfig
	flags                     = *flag.NewFlagSet("peereval-plugin", flag.ContinueOnError)
	maxPeerCount              = flags.Int("peereval.max.peers", 0, "max peer value for peer eval plugin")
	pollingInterval           = flags.Duration("peereval.polling.interval", time.Minute, "polling interval for peer monitoring")
	connectionTimeCoefficient = flags.Duration("peereval.connection.time.coefficient", 5*time.Minute, "minimum connection time before evaluating peers")
)

type peerEvalModule struct {
	peerMetricsMap map[string]*PeerMetrics
	mutex          sync.Mutex
	peerRatios     map[string]float64
}

func init() {
	xplugeth.RegisterModule[peerEvalModule]("peerEvalModule")
	xplugeth.RegisterHook[HealthyPeers]()
	xplugeth.RegisterFlags(flags)
}

func (p *peerEvalModule) InitializeNode(s *node.Node, b types.Backend) {
	client =  s.Attach()

	config, ok := xplugeth.GetConfig[peerEvalConfig]("peereval")
	if !ok {
		log.Warn("peerEval config not found, using defaults")
		cfg = &peerEvalConfig{}
	} else {
		cfg = config
	}

	if *maxPeerCount == 0 {
		*maxPeerCount = 3
		log.Warn(fmt.Sprintf("max peer count flag not set, peer eval plugin, setting to a default of %v", *maxPeerCount))
	}
	p.peerMetricsMap = make(map[string]*PeerMetrics)
	p.StartPeerMonitoring()
	p.cleanUpPeerMap()

	log.Info("Initialized node, peer eval plugin")
	log.Info(fmt.Sprintf("Polling interval set to %v minutes", pollingInterval.Minutes()))
}

func (p *peerEvalModule) Blockchain() {
	go p.streamHealthyPeers()
}

type peerInfo struct {
    ID      string
	Enode   string 
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
			for id, peer := range p.peerMetricsMap {
				connectedDuration := peer.ConnectedTime + time.Since(peer.LastConnected)
				if !peer.IsConnected && connectedDuration >= *connectionTimeCoefficient {
					delete(p.peerMetricsMap, id)
				}
			}
		}
	}()
}

func (p *peerEvalModule) updatePeerConnections() {
	log.Error("inside update peerconnections")
	p.mutex.Lock()
	defer p.mutex.Unlock()

	peers, err := getPeers()
	if err != nil {
		log.Error("Failed to get peers", "err", err)
		return
	}

    log.Error("length of peerMetricsMap", "length", len(p.peerMetricsMap))

	currentPeers := make(map[string]bool)
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

	p.peerRatios = make(map[string]float64)
		for id, peer := range p.peerMetricsMap {
			if peer.IsConnected {
				connectedDuration := peer.ConnectedTime + time.Since(peer.LastConnected)
				if connectedDuration >= *connectionTimeCoefficient {
					p.peerRatios[id] = float64(peer.BlocksContributed) / connectedDuration.Seconds()
			}
		}
	}

	if len(p.peerMetricsMap) > int(float64(*maxPeerCount)*0.9) {
		p.prunePeers(true)
	} else {
		var outboundCount float64
		for _, peer := range p.peerMetricsMap{
			if !peer.IsInbound{
				outboundCount++
			}
		}

		if outboundCount >= float64(len(p.peerMetricsMap)) * 0.9 {
			p.prunePeers(false)
		}
	}
}

func (p *peerEvalModule) GetHealthyPeers() []string{
	p.mutex.Lock()
	defer p.mutex.Unlock()

	var healthy []string
	for _, metrics := range p.peerMetricsMap {
		if metrics.BlocksContributed > 0 && metrics.Enode != "" {
			healthy = append(healthy, metrics.Enode)
		}
	}
	return healthy
}

func (p *peerEvalModule) streamHealthyPeers() {
	ticker := time.NewTicker(*pollingInterval)
	defer ticker.Stop()

	time.Sleep(*pollingInterval)

	producer, err := utils.CreateProducer(cfg.BrokerURL, cfg.Topic)
	if err != nil {
		log.Error("failed to create Kafka producer", "err", err)
		return
	}

	for range ticker.C {
		peers := p.GetHealthyPeers()
		log.Error("len of healthyPeers", "length", len(peers))

		if len(peers) == 0 {
			continue
		}

		data, err := json.Marshal(peers)
		if err != nil {
			log.Error("failed to marshal healthy peers", "err", err)
			continue
		}
		msg := &sarama.ProducerMessage{
			Topic : cfg.Topic,
			Value : sarama.ByteEncoder(data),
		}
		log.Error("sending getHealthyPeers")
		producer.Input() <-msg 
	}
}

func (p *peerEvalModule) prunePeers(pruneAll bool){
	if len(p.peerRatios) == 0 {
		return
	}
	var peersToDrop []string
	for id, peer := range p.peerMetricsMap {
		if pruneAll || !peer.IsInbound {
			peersToDrop = append(peersToDrop, id)
		}
	}
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
	_ initialize.Blockchain  = (*peerEvalModule)(nil)
	_ initialize.Initializer = (*peerEvalModule)(nil)
	_ fetcher.PeerEvalPlugin = (*peerEvalModule)(nil)
)
