package forkready

import (
	"reflect"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/openrelayxyz/xplugeth"
	"github.com/openrelayxyz/xplugeth/hooks/apis"
	"github.com/openrelayxyz/xplugeth/hooks/initialize"
	"github.com/openrelayxyz/xplugeth/types"
)

var chainConfig any

type forkReadyModule struct {
}

func init() {
	xplugeth.RegisterModule[forkReadyModule]("forkReady")
}

func (r *forkReadyModule) InitializeNode(stack *node.Node, backend types.Backend, cfg any) {
	chainConfig = cfg
	log.Info("forkReady module initialized")
}

func (r *forkReadyModule) GetAPIs(s *node.Node, b types.Backend) []rpc.API {
	return []rpc.API{
		{
			Namespace: "cardinal",
			Service: &forkReadyAPI{
				stack: s,
			},
		},
	}
}

type forkReadyAPI struct{
	stack *node.Node
}

func (r *forkReadyAPI) ForkReady(forkName string) int {
	result := -1
	client := r.stack.Attach()
	if client == nil {
		log.Error("error acquiring client, ForkReady plugin")
		return result
	}
	switch forkName {
		case "osaka":
			var latestBlock map[string]interface{}
			if err := client.Call(&latestBlock, "eth_getBlockByNumber", "latest", false); err != nil {
				log.Error("error returned from client call, forkReady plugin", "err", err)
				return result
			}
			blockTime, err := hexutil.DecodeUint64(latestBlock["timestamp"].(string))
			if err != nil {
				log.Error("error decoding latest block timestamp, forkReady plugin", "err", err)
				return result
			}
			metaCfg := reflect.ValueOf(chainConfig).Type().Elem()
			if _, ok := metaCfg.FieldByName("OsakaTime"); ok {
				ptr := chainConfig.(*params.ChainConfig)
				cfg := *ptr
				if oTime := cfg.OsakaTime; oTime != nil {
					if blockTime >= *oTime {
						result = 2
					} else {
						result = 1
					} 
				} else {
					result = 0
				}
			}
	}

	return result
}

var (
	_ apis.GetAPIs = (*forkReadyModule)(nil)
	_ initialize.Initializer = (*forkReadyModule)(nil)
)