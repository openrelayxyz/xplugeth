package merge

import (
	"fmt"
	"math/big"

	glog "github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	ctypes "github.com/openrelayxyz/cardinal-types"
	"github.com/openrelayxyz/cardinal-types/hexutil"
	"github.com/openrelayxyz/cardinal-types/metrics"
	"github.com/openrelayxyz/xplugeth"
	"github.com/openrelayxyz/xplugeth/types"
)

type mergePlugin struct{}

var (
	postMerge       bool
	backend         types.Backend
	gethWeightGauge = metrics.NewMajorGauge("/geth/weight")
	stack           node.Node
	chainid         int64
	log             glog.Logger
)

type numLookup struct {
	Number hexutil.Big `json:"number"`
}

func init() {
	xplugeth.RegisterModule[mergePlugin]()
}

func (*mergePlugin) InitializeNode(s *node.Node, b types.Backend) {
	stack = *s
	backend = b
	chainid = b.ChainConfig().ChainID.Int64()
}

func getSafeFinalized() (*big.Int, *big.Int) {
	client := stack.Attach()
	var snl, fnl numLookup
	if err := client.Call(&snl, "eth_getBlockByNumber", "safe", false); err != nil {
		log.Warn("Could not get safe block", "err", err)
	}
	if err := client.Call(&fnl, "eth_getBlockByNumber", "finalized", false); err != nil {
		log.Warn("Could not get finalized block", "err", err)
	}
	return snl.Number.ToInt(), fnl.Number.ToInt()
}

func CardinalAddBlockHook(number int64, hash, parent ctypes.Hash, weight *big.Int, updates map[string][]byte, deletes map[string]struct{}) {
	log.Info("add cardinalBlockHook")
	if !postMerge {
		v, _ := backend.ChainDb().Get([]byte("eth2-transition"))
		if len(v) > 0 {
			postMerge = true
		} else {
			// Not yet post merge, we don't want to make any modifications
			gethWeightGauge.Update(new(big.Int).Div(weight, big.NewInt(10000000000000000)).Int64())
			return
		}
	}
	snum, fnum := getSafeFinalized()
	if snum != nil {
		updates[fmt.Sprintf("c/%x/n/safe", chainid)] = snum.Bytes()
	}
	if fnum != nil {
		updates[fmt.Sprintf("c/%x/n/finalized", chainid)] = fnum.Bytes()
	}
	// After the merge, the td of a block stops increasing, but certain elements
	// of  Cardinal still needs a weight for evaluating block ordering. The
	// convention for this is to add the block number to the final total
	// difficulty to choose a weight.
	weight.Add(weight, big.NewInt(number))
	gethWeightGauge.Update(new(big.Int).Div(weight, big.NewInt(10000000000000000)).Int64())
}
