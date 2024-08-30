package trieflushintervalclone

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/openrelayxyz/xplugeth"
	"github.com/openrelayxyz/xplugeth/types"
)

type trieflushModule struct{}

func init() {
	xplugeth.RegisterModule[trieflushModule]()
}

func (*trieflushModule) InitializeNode(s *node.Node, b types.Backend) {
	log.Info("trieflushinterval plugin initialized")
}

func (*trieflushModule) GetAPIs(stack *node.Node, backend types.Backend) []rpc.API {
	return []rpc.API{
		{
			Namespace: "debug",
			Version:   "1.0",
			Service:   &trieflushModule{},
			Public:    true,
		},
	}
}

var nodeInterval time.Duration

var ModifiedInterval time.Duration

func (*trieflushModule) SetTrieFlushIntervalClone(duration time.Duration) time.Duration {
	nodeInterval = duration
	if ModifiedInterval > 0 {
		duration = ModifiedInterval
	}
	return duration
}

func (*trieflushModule) SetTrieFlushInterval(ctx context.Context, interval string) error {
	newInterval, err := time.ParseDuration(interval)
	if err != nil {
		return err
	}
	ModifiedInterval = newInterval

	return nil
}

func (*trieflushModule) Test(context.Context) string {
	return "test successful"
}
