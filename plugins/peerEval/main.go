package peereval

import (
	"fmt"
	"time"

	gtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/rpc"
	
	"github.com/openrelayxyz/xplugeth"
	"github.com/openrelayxyz/xplugeth/hooks/apis"
	"github.com/openrelayxyz/xplugeth/hooks/blockchain"
	"github.com/openrelayxyz/xplugeth/hooks/initialize"
	"github.com/openrelayxyz/xplugeth/types"
)

var (
	stack       node.Node
	client      *rpc.Client
	gatheringCount 		int
	innerPeerData    []map[string]interface{}
	outerPeerData    map[string]interface{}
)

type peerEvalPlugin struct {
}

func init() {
	xplugeth.RegisterModule[peerEvalPlugin]("peerEvalPlugin")
}

func (p *peerEvalPlugin) InitializeNode(s *node.Node, _ types.Backend) {
	client = s.Attach()
}

func getPeers() ([]string, error) {
	var rawPeerData []map[string]interface{}
	err := client.Call(&rawPeerData, "admin_peers")
	if err != nil {
		log.Error("error calling admin_peers, peerEval plugin", "err", err)
		return nil, err
	}

	peers := []string{}

	for _, item := range rawPeerData {
		for k, v := range item {
			if k == "id" {
				peers = append(peers, v.(string))
			}
		}
	}

	return peers, nil

}

func (p *peerEvalPlugin) PeerEval(id string, headers []*gtypes.Header) {
	gatheringCount ++
	log.Error(fmt.Sprintf("Gathering peer data, count %v/100", gatheringCount))

	t := time.Now().Format("20060102_150405")

	blockNumbers := []string{}
	for _, header := range headers {
		blockNumbers = append(blockNumbers, header.Number.String())
	}

	evalData := map[string]interface{}{
			"id":     id,
			"time":   t,
			"blocks": blockNumbers,
	}

	innerPeerData = append(innerPeerData, evalData)

	if gatheringCount >= 100 {
		gatheringCount = 0
	}

}

type peerEvalAPI struct {}

func (p *peerEvalAPI) GetPeerData() (map[string]interface{}, error) {
	resultChan := make(chan map[string]interface{}, 1)
	errChan := make(chan error, 1)
	defer close(resultChan)
	defer close(errChan)
	go func() {
		for {
			if gatheringCount >= 100 {
				peerSlice, err := getPeers()
				if err != nil {
					log.Error("error obtaining peer slice", "err", err)
					errChan <- err

				} 
				data := make(map[string]interface{})
				data["peers"] = peerSlice
				data["active"] = innerPeerData
				
				resultChan <- data
			}
		}
	}()

	// var err error
	// var result map[string]interface{}
	select {
	case err := <-errChan:
		return nil, err
	case result := <-resultChan:
		return result, nil
	}
}

func (p *peerEvalPlugin) GetAPIs(*node.Node, types.Backend) []rpc.API {
	log.Info("Registering peer eval plugin APIs")
	return []rpc.API{
		{
			Namespace: "plugeth",
			Service:   &peerEvalAPI{},
		},
	}
}

var (
	_ apis.GetAPIs    = (*peerEvalPlugin)(nil)
	_ initialize.Initializer    = (*peerEvalPlugin)(nil)
	_ blockchain.PeerEvalPlugin = (*peerEvalPlugin)(nil)
)
