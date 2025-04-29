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
	"encoding/json"

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


var (
	_ initialize.Blockchain  = (*peerManagerModule)(nil)
	_ initialize.Initializer = (*peerManagerModule)(nil)
)
