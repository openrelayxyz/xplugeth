package forkready

import (
	"reflect"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/openrelayxyz/xplugeth"
	"github.com/openrelayxyz/xplugeth/hooks/apis"
	"github.com/openrelayxyz/xplugeth/hooks/initialize"
	"github.com/openrelayxyz/xplugeth/types"
)

type forkReadyModule struct {
}

func init() {
	xplugeth.RegisterModule[forkReadyModule]("forkReady")
}

func (r *forkReadyModule) InitializeNode(stack *node.Node, backend types.Backend, cfg any) {
	log.Info("forkReady module initialized")
}

func (r *forkReadyModule) GetAPIs(s *node.Node, b types.Backend, c any) []rpc.API {
	return []rpc.API{
		{
			Namespace: "cardinal",
			Service: &forkReadyAPI{
				stack: s,
				chainConfig: c,
			},
		},
	}
}

type forkReadyAPI struct{
	stack *node.Node
	chainConfig any
}

type psudoConfig struct {
	OsakaTime *uint64
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
			metaCfg := reflect.ValueOf(r.chainConfig)
			if metaCfg.Kind() == reflect.Ptr {
				metaCfg = metaCfg.Elem()
			}
			val := metaCfg.FieldByName("OsakaTime") 
			if val.IsValid() {
				if !val.IsNil() {
					oTime := *val.Interface().(*uint64)
					// this may need to be complexified if / when other supported chains add osakaTime to their chain configs
					if blockTime >= oTime {
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