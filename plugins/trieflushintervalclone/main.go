package trieflushintervalclone

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/openrelayxyz/xplugeth"
	"github.com/openrelayxyz/xplugeth/types"
	"github.com/openrelayxyz/xplugeth/hooks/apis"
	"github.com/openrelayxyz/xplugeth/hooks/blockchain"
)

type trieFlushModule struct{}

type trieFlushAPI struct {}

func init() {
	xplugeth.RegisterModule[trieFlushModule]("trieFlushClone")
}

func (*trieFlushModule) GetAPIs(stack *node.Node, backend types.Backend) []rpc.API {
	return []rpc.API{
		{
			Namespace: "debug",
			Version:   "1.0",
			Service:   &trieFlushAPI{},
			Public:    true,
		},
	}
}

var (
	nodeInterval time.Duration
	ModifiedInterval time.Duration
)

func (*trieFlushModule) SetTrieFlushIntervalClone(duration time.Duration) time.Duration {
	nodeInterval = duration
	if ModifiedInterval > 0 {
		duration = ModifiedInterval
	}
	return duration
}

func (*trieFlushAPI) SetTrieFlushInterval(ctx context.Context, interval string) error {
	newInterval, err := time.ParseDuration(interval)
	if err != nil {
		return err
	}
	log.Info("flush interval set from plugin", "interval", interval)
	ModifiedInterval = newInterval

	return nil
}

var (
	_ apis.GetAPIs = (*trieFlushModule)(nil)
	_ blockchain.SetTrieFlushIntervalClonePlugin = (*trieFlushModule)(nil)
)
