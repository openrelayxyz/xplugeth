package hooktest

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/openrelayxyz/xplugeth/types"
	"github.com/openrelayxyz/xplugeth/hooks/blockchain"
)


func (p *hookTest) SetTrieFlushIntervalClone(duration time.Duration) time.Duration {
	nodeInterval = duration

	if modifiedInterval > 0 {
		duration = modifiedInterval
	}

	return duration
}

type hookTestAPI struct {}

func (p *hookTestAPI) SetTrieFlushInterval(ctx context.Context, interval string) error {
	newInterval, err := time.ParseDuration(interval)
	if err != nil {
		return err
	}
	modifiedInterval = newInterval

	return nil
}

func (p *hookTest) GetAPIs(stack *node.Node, backend types.Backend) []rpc.API {
	return []rpc.API{
		{
			Namespace: "debug",
			Version:   "1.0",
			Service:   &hookTestAPI{},
			Public:    true,
		},
	}
}

var (
	_ blockchain.SetTrieFlushIntervalClonePlugin = (*hookTest)(nil)
)