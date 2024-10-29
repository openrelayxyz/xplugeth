package peermanager

import (
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/openrelayxyz/xplugeth"
	"github.com/openrelayxyz/xplugeth/hooks/initialize"
	"github.com/openrelayxyz/xplugeth/types"
)

type peerManagerConfig struct {
	BrokerURL string `yaml:"peerManager.broker.url"`
}

var cfg *peerManagerConfig

var (
	httpApiFlagName    = "http.api"
	sessionStack       node.Node
	sessionBrokers     []string
	sessionKafkaConfig *sarama.Config
	nodes              = make(chan string, 5)
	exit               = make(chan struct{}, 1)
	peerBroker         string
)

type peerManagerModule struct{}

func init() {
	xplugeth.RegisterModule[peerManagerModule]("peerManagerModule")
}

func (*peerManagerModule) InitializeNode(s *node.Node, b types.Backend) {
	sessionStack = *s
	var ok bool
	cfg, ok = xplugeth.GetConfig[peerManagerConfig]("peerManager")
	peerBroker = cfg.BrokerURL
	if !ok {
		log.Warn("did not acqire config, example plugin, all values set to default")
	}
	log.Info("Initialized node, peer manager plugin")
}

func (*peerManagerModule) Blockchain() {
	if peerBroker == "" {
		panic(fmt.Sprintf("no broker provided for peer manager plugin"))
	}
	go peeringSequence()
}

func peeringSequence() {
	sessionPeerService, err := getPeerManager()
	if err != nil {
		log.Error("session peer service unavailable, peer manager plugin", "err", err)
		return
	}

	selfNode, err := sessionPeerService.getEnode()
	if err != nil {
		log.Error("error calling getEnode from sessionService, peer manager plugin", "err", err)
	}

	chainTopic, err := sessionPeerService.chainIdResolver()
	if err != nil {
		log.Error("Error aquiring chainID, peer manager plugin", "err", err)
	}

	producer, err := createProducer(peerBroker, chainTopic)
	if err != nil {
		log.Error("failed to acquire kafka producer, peer manager plugin", "err", err)
		return
	}

	consumer, err := createConsumer(peerBroker, chainTopic)
	if err != nil {
		log.Error("failed to acquire kafka consumer, peer manager plugin", "err", err)
		return
	}

	msg := &sarama.ProducerMessage{
		Topic: chainTopic,
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
