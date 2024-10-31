package peermanager

import (
	"fmt"
	"net/http"
	"encoding/json"
	"io/ioutil"
	"strings"
	"time"

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

	var	myip string
	for {
		if myip = getPublicIP(); myip != "" {
			break
		}
		time.Sleep(1 * time.Second)
	}

	var raw pseudoNodeInfo
	err := service.client.Call(&raw, "admin_nodeInfo")
	if err != nil {
		return "", err
	}
	parts := strings.Split(raw.Enode, "@")
	port := strings.Split(parts[1], ":")[1]

	return parts[0] + "@" + myip + ":" + port, nil
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

type myIp struct {
	IP string `json:"ip"`
}

func getPublicIP() string {
	resp, err := http.Get("https://myipv4.p1.opendns.com/get_my_ip")
	if err != nil {
		log.Error("error retrieving ip, peer manager, retrying", "err", err)
		return ""
	}
	defer resp.Body.Close()

	raw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("error reading ip response, peer manager, retrying", "err", err)
		return ""
	}

	var myip myIp 
	if err := json.Unmarshal(raw, &myip); err != nil {
		log.Error("error unmarshaling myip json, peer manager, retrying", "err", err)
		return ""
	} 

	return myip.IP
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
