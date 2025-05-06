package peermanager

import (
	"fmt"

	"encoding/json"

	"github.com/Shopify/sarama"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"

	"github.com/openrelayxyz/xplugeth"
	"github.com/openrelayxyz/xplugeth/hooks/initialize"
	"github.com/openrelayxyz/xplugeth/types"
	"github.com/openrelayxyz/xplugeth/utils"

	"github.com/openrelayxyz/xplugeth/plugins/peereval"
	_ "github.com/openrelayxyz/xplugeth/plugins/peereval"
)

type peerManagerConfig struct {
	BrokerURL string `yaml:"broker.url"`
	PeerTopic string `yaml:"peer.topic"`
}

var (
	sessionPeerService *PeerManager
	chainid            int64
	brokers            []string
	config             *sarama.Config
	nodes              = make(chan string, 5)
	exit               = make(chan struct{}, 1)
	cfg                *peerManagerConfig
	peerBroker         string
	peerTopic          string
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
		log.Warn("did not acqire config, example plugin, all values set to default")
	}
	peerBroker = cfg.BrokerURL
	peerTopic = cfg.PeerTopic

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

	type peerBroadcast struct {
		Generic []string `json:"generic"`
	}

	var eval []peereval.HealthyPeers
	if xplugeth.HasModule("peerEvalModule") {
		eval = xplugeth.GetModules[peereval.HealthyPeers]()
		if len(eval) == 0 {
            log.Warn("peerEvalModule present but no GetHealthyPeers found")
        }
	}

	if len(eval) > 0 {
		peers := eval[0].GetHealthyPeers()
		payload := peerBroadcast{
			Generic: peers,
		}
		data, err := json.Marshal(payload)
		if err != nil {
			log.Error("failed to marshal peer payload", "err", err)
			return
		}
		msg := &sarama.ProducerMessage{
				Topic: peerTopic,
				Value: sarama.ByteEncoder(data),
		}

		producer.Input() <- msg

		go func(){
			for message := range consumer.Messages() {
				var incoming peerBroadcast
				if err := json.Unmarshal(message.Value, &incoming); err != nil{
					log.Error("failed to unmarshal peer payload", "err", err)
					continue
				}
				for _, node := range incoming.Generic {
					nodes <- node
				}
			}
		}()

		for _, n := range payload.Generic{
			if n == selfNode{
				continue
			} else {
				sessionPeerService.attachPeerOnly(n)
			}
		}

		for node := range nodes {
			if node == selfNode {
				continue
			} else {
				sessionPeerService.attachPeers(node)
			}
		}
	} else{
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
}


var (
	_ initialize.Blockchain  = (*peerManagerModule)(nil)
	_ initialize.Initializer = (*peerManagerModule)(nil)
)
