package peermanager

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

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
	var pni pseudoNodeInfo
	err := service.client.Call(&pni, "admin_nodeInfo")
	if err != nil {
		return "", err
	}

	return analyzeEnode(pni.Enode), nil
}

func (service *PeerManager) attachTrustedPeer(peer string) error {
	if err := service.attachPeer(peer); err != nil {
		return err
	}
	var result bool
	if err := service.client.Call(&result, "admin_addTrustedPeer", peer); err != nil {
		return err
	}
	return nil
}

func (service *PeerManager) attachPeer(peer string) error {
	var result bool
	if err := service.client.Call(&result, "admin_addPeer", peer); err != nil {
		return err
	}
	return nil
}

type myIp struct {
	IP string `json:"ip"`
}

func analyzeEnode(raw string) string {
	firstPass := strings.Split(raw, "@")
	nodeId := firstPass[0]

	secondPass := strings.Split(firstPass[1], ":")
	ip := secondPass[0]
	port := secondPass[1]

	if ip == "127.0.0.1" {
		if publicIP := getPublicIP(); publicIP != "" {
			ip = publicIP
		}
	}

	return nodeId + "@" + ip + ":" + port
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
