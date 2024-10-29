package peermanager

import (
	"fmt"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
)

type PeerManagerService struct {
	client *rpc.Client
}

type pseudoNodeInfo struct {
	Enode string
}

func getPeerManager() (*PeerManagerService, error) {
	c := sessionStack.Attach()

	s := &PeerManagerService{
		client: c,
	}
	return s, nil
}

func (service *PeerManagerService) getEnode() (string, error) {
	var enode pseudoNodeInfo
	err := service.client.Call(&enode, "admin_nodeInfo")
	if err != nil {
		return "", err
	}
	return enode.Enode, nil
}

func (service *PeerManagerService) attachPeers(peer string) {

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

func (service *PeerManagerService) chainIdResolver() (string, error) {

	var chainID string
	err := service.client.Call(&chainID, "eth_chainId")
	if err != nil {
		return "", err
	}
	var result string
	switch chainID {
	case "0x1":
		result = "mainnet"
	case "0x3d":
		result = "etc"
	case "0x4268":
		result = "holesky"
	case "0xaa36a7":
		result = "sepolia"
	case "0x89":
		result = "polygon"
	case "0x13881":
		result = "mumbai"
	default:
		log.Warn("unknown chain, chainID could not be resolved, peer manager plugin")
		result = chainID
	}
	return fmt.Sprintf("peers-%v", result), nil
}
