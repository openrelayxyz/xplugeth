package peermanager

import (
	"fmt"
	"strings"

	"encoding/json"

	"github.com/Shopify/sarama"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"

	"github.com/openrelayxyz/xplugeth"
	"github.com/openrelayxyz/xplugeth/hooks/initialize"
	"github.com/openrelayxyz/xplugeth/types"
	xp-utils "github.com/openrelayxyz/xplugeth/utils"
)

type peerBroadcast struct {
	Trusted string `json:"trusted"`
	Generic []string `json:"generic"`
}
type peerManagerConfig struct {
	BrokerURL string `yaml:"broker.url"`
	PeerTopic string `yaml:"peer.topic"`
	EvalTopic string `yaml:"eval.topic"`
}

var (
	sessionPeerService *PeerManager
	chainid            int64
	nodes              = make(chan string, 5)
	exit               = make(chan struct{}, 1)
	cfg                *peerManagerConfig
)

type peerManagerModule struct {
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
		log.Warn("did not acqire config, peermanager plugin, all values set to default")
	}

	log.Info("Initialized node, peer manager plugin")
}

func (p *peerManagerModule) Blockchain() {
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

	producer, err :=  xp-utils.CreateProducer(cfg.BrokerURL, cfg.PeerTopic)
	if err != nil {
		log.Error("failed to acquire kafka producer, peer manager plugin", "err", err)
		return
	}

	consumer, err := xp-utils.CreateConsumer(cfg.BrokerURL, cfg.PeerTopic)
	if err != nil {
		log.Error("failed to acquire kafka consumer, peer manager plugin", "err", err)
		return
	}

	if xplugeth.HasModule("peerEvalModule") {
		log.Error("peereval module present")
		evalConsumer, err := xp-utils.CreateConsumer(cfg.BrokerURL, cfg.EvalTopic)
		if err != nil {
			log.Error("failed to acquire peereval consumer, peer manager plugin", "err", err)
			return
		}

		initialPayload := peerBroadcast{
			Trusted: selfNode,
			Generic: nil,
		}
		data, err := json.Marshal(initialPayload)
		if err == nil {
			msg := &sarama.ProducerMessage{
				Topic: cfg.PeerTopic,
				Value: sarama.ByteEncoder(data),
			}
			log.Error("sending initial message", "enode", string(data))
			producer.Input() <- msg
		}

		go func() { 
			for message := range evalConsumer.Messages(){
				var genericPeers []string
				if err := json.Unmarshal(message.Value, &genericPeers); err != nil {
					log.Error("failed to unmarshal peerEval payload", "err", err)
					continue
				}
				
				payload := &peerBroadcast{
					Generic: genericPeers,
				}

				data, err := json.Marshal(payload);
				if err != nil {
					log.Error("failed to marshal peerBroadcast", "err", err)
				}

				peerMsg := &sarama.ProducerMessage{
					Topic: cfg.PeerTopic,
					Value : sarama.ByteEncoder(data),
				}
				producer.Input() <- peerMsg
			}
		}()

		go func() {
			for message := range consumer.Messages(){
				var incoming peerBroadcast
				if err := json.Unmarshal(message.Value, &incoming); err != nil {
					log.Error("failed to unmarshal peer broadcast", "err", err)
					continue
				}

				if incoming.Trusted != "" {
					if !isPeerConnected(incoming.Trusted) && incoming.Trusted != selfNode {
						sessionPeerService.attachTrustedPeer(incoming.Trusted)
					}else {
						log.Error("skipping trusted cause it already exists", "peer", incoming.Trusted)
					}
				} 

				for _, peer := range incoming.Generic{
					if peer != selfNode && !isPeerConnected(peer){
						sessionPeerService.attachPeerOnly(peer)
					} else {
						log.Error("skipping generic cause it already exists")
					}
				}
			}
		}()

	} else {
		msg := &sarama.ProducerMessage{
			Topic: cfg.PeerTopic,
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
				log.Error("received own node message", "node", selfNode)
				continue
			} else {
				sessionPeerService.attachTrustedPeer(message)
			}
		}
	}
}

func isPeerConnected (enode string) bool {
	var peerList []map[string]interface{}
	err := sessionPeerService.client.Call(&peerList, "admin_peers")
	if err != nil {
		log.Error("error calling admin_peers", "err", err)
		return false
	}
	for _, peer := range peerList{
		if id, ok := peer["id"].(string); ok && strings.Contains(enode, id) {
			return true
		}
	}
	return false
}

var (
	_ initialize.Blockchain  = (*peerManagerModule)(nil)
	_ initialize.Initializer = (*peerManagerModule)(nil)
)
