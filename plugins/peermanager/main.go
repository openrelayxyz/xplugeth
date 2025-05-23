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

type PeerManagerConfig struct {
	BrokerURL string `yaml:"broker.url"`
	PeerTopic string `yaml:"peer.topic"`
}

var (
	sessionPeerService *PeerManager
	chainid            int64
	nodes              = make(chan string, 5)
	exit               = make(chan struct{}, 1)
	SharedTopic        *string
	SharedBroker       *string
)

type peerManagerModule struct {
	producer sarama.AsyncProducer
	consumer sarama.PartitionConsumer
	selfNode string
	cfg      *PeerManagerConfig
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

	cfg, ok := xplugeth.GetConfig[PeerManagerConfig]("peermanager")
	if !ok {
		log.Warn("did not acqire config, peermanager plugin, peering sequence unavailable")
		return
	} else {
		p.cfg = cfg

		SharedBroker =  &cfg.BrokerURL
		SharedTopic = &cfg.PeerTopic

		selfNode, err := sessionPeerService.getEnode()
		if err != nil {
			log.Error("error calling getEnode from sessionService, peer manager plugin", "err", err)
			return
		}
		p.selfNode = selfNode

		producer, err :=  xp_utils.CreateProducer(cfg.BrokerURL, cfg.PeerTopic)
		if err != nil {
			log.Error("failed to acquire kafka producer, peer manager plugin", "err", err)
			return
		}

		p.producer = producer

		consumer, err := xp_utils.CreateConsumer(cfg.BrokerURL, cfg.PeerTopic)
		if err != nil {
			log.Error("failed to acquire kafka consumer, peer manager plugin", "err", err)
			return
		}

		p.consumer = consumer

		go p.peeringSequence()
	}

	log.Info("Initialized node, peer manager plugin")
}

func (p *peerManagerModule) peeringSequence() {

	go func ()  {
		ticker := time.NewTicker(20 * time.Minute)
		defer ticker.Stop()

		if err := p.broadcastSelfNode(); err != nil {
			log.Error("failed to broadcase selfnode peermanager", "err", err)
		}

		for range ticker.C {
			if err := p.broadcastSelfNode(); err != nil {
				log.Error("failed to broadcase selfnode peermanager", "err", err)
			}
		}	
	}()

	go func() {
		for message := range p.consumer.Messages() {
			var incoming peerBroadcast
			if err := json.Unmarshal(message.Value, &incoming); err != nil {
				log.Error("failed to unmarshal peer broadcast", "err", err)
				continue
			}
			log.Error("we made it here??", "len", len(incoming.peers))
			if len(incoming.peers) > 0 {
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
		}
	}()
}

func (p *peerManagerModule) broadcastSelfNode() error {
	payload := &peerBroadcast{
		peers: []string{p.selfNode},
	}
	data, err := json.Marshal(payload)
	if err != nil {
		log.Error("failed to marshal peerBroadcast, default peermanager", "err", err)
		return err
		
	}

	msg := &sarama.ProducerMessage{
		Topic: p.cfg.PeerTopic,
		Value: sarama.ByteEncoder(data),
	}
	// var check peerBroadcast
	// if err := json.Unmarshal(msg, &check); err != nil {
	// 	log.Error("failed to unmarshal peer broadcast", "err", err)
	// }
	log.Error("broadcast self node", "enode", p.selfNode, "msg", msg, "payload", payload)
	p.producer.Input() <- msg
	return nil
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
