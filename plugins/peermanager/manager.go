package peermanager

import (
	"fmt"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
)

type pseudoNodeInfo struct {
	Enode string `json:"enode"`
}

type PeerManager struct {
	client *rpc.Client
}

func (service *PeerManager) getEnode() (string, error) {
	var enode pseudoNodeInfo
	err := service.client.Call(&enode, "admin_nodeInfo")
	if err != nil {
		return "", err
	}
	return enode.Enode, nil
}

func (service *PeerManager) attachPeers(peer string) {

	var addTrustedPeerResult bool
	err := service.client.Call(&addTrustedPeerResult, "admin_addTrustedPeer", peer)
	if err != nil {
		log.Error("error calling admin_addTrustedPeer, peer manager plugin", "peer", peer, "err", err)
	}
	if !addTrustedPeerResult {
		log.Error("addTrustedPeer returned false, peer manager plugin", "peer", peer, "err", err)
	}

	var addPeerResult bool
	err = service.client.Call(&addPeerResult, "admin_addPeer", peer)
	if err != nil {
		log.Error("error calling admin_addPeer, peer manager plugin", "peer", peer, "err", err)
	}
	if !addPeerResult {
		log.Error("addPeer returned false, peer manager plugin", "peer", peer, "err", err)
	}
	log.Info("added peer, peer manager plugin", "added", peer)
}

func chainIdResolver(id int64) string {
	var result string
	switch id {
	case 1:
		result = "mainnet"
	case 61:
		result = "etc"
	case 17000:
		result = "holesky"
	case 11155111:
		result = "sepolia"
	case 137:
		result = "polygon"
	case 80001:
		result = "mumbai"
	case 80002:
		result = "amoy"
	default:
		log.Warn("unknown chain, chainID could not be resolved, peer manager plugin")
		result = fmt.Sprintf("%x", id)
	}
	return fmt.Sprintf("peers-%v", result)
}
