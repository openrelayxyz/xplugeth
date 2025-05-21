package peermanager

import (
	"fmt"
	"strings"
	"time"

	"encoding/json"

	"github.com/Shopify/sarama"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"

	"github.com/openrelayxyz/xplugeth"
	"github.com/openrelayxyz/xplugeth/hooks/initialize"
	"github.com/openrelayxyz/xplugeth/types"
	"github.com/openrelayxyz/xplugeth/utils"
	xp_utils "github.com/openrelayxyz/xplugeth/utils/plugins"
)

type peerBroadcast struct {
	peers []string
}

type peerManagerConfig struct {
	brokerURL string `yaml:"broker.url"`
	peerTopic string `yaml:"peer.topic"`
}

var (
	sessionPeerService *PeerManager
	chainid            int64
	nodes              = make(chan string, 5)
	exit               = make(chan struct{}, 1)
	cfg                *peerManagerConfig
	SharedTopic        *string
	SharedBroker       *string
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
		log.Warn("did not acqire config, peermanager plugin, peering sequence unavailable")
		return
	} else {
		SharedBroker =  &cfg.brokerURL
		SharedTopic = &cfg.peerTopic
		go peeringSequence()
	}

	log.Info("Initialized node, peer manager plugin")
}

func peeringSequence() {
	selfNode, err := sessionPeerService.getEnode()
	if err != nil {
		log.Error("error calling getEnode from sessionService, peer manager plugin", "err", err)
	}

	producer, err :=  xp_utils.CreateProducer(cfg.brokerURL, cfg.peerTopic)
	if err != nil {
		log.Error("failed to acquire kafka producer, peer manager plugin", "err", err)
		return
	}

	consumer, err := xp_utils.CreateConsumer(cfg.brokerURL, cfg.peerTopic)
	if err != nil {
		log.Error("failed to acquire kafka consumer, peer manager plugin", "err", err)
		return
	}

	go func ()  {
		ticker := time.NewTicker(20 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			payload := &peerBroadcast{
				peers: []string{selfNode},
			}
			data, err := json.Marshal(payload)
			if err != nil {
				log.Error("failed to marshal peerBroadcast, default peermanager", "err", err)
				continue
			}

			msg := &sarama.ProducerMessage{
				Topic: cfg.peerTopic,
				Value: sarama.ByteEncoder(data),
			}
			producer.Input() <- msg
			log.Error("broadcast self node", "enode", selfNode)
		}	
	}()

	go func() {
		for message := range consumer.Messages() {
			var incoming peerBroadcast
			if err := json.Unmarshal(message.Value, &incoming); err != nil {
				log.Error("failed to unmarshal peer broadcast", "err", err)
				continue
			}
			if incoming.peers[0] != "" && !isPeerConnected(incoming.peers[0]) {
				if err := sessionPeerService.attachTrustedPeer(incoming.peers[0]); err != nil {
					log.Error("error attaching trusted peer, peermanager", "trusted peer", incoming.peers[0], "err", err)
				}
				log.Error("**** Added trusted peer ****", "peer", incoming.peers[0])
			}
			for _, peer := range incoming.peers[1:] {
				if !isPeerConnected(peer) {
					if err := sessionPeerService.attachPeer(peer); err != nil {
						log.Error("error attaching generic peer, peermanager", "peer", peer, "err", err)
					}
					log.Error("**** Added generic peer ****", "peer", peer)
				}
			}
		}
	}()
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
	_ initialize.Initializer = (*peerManagerModule)(nil)
)
