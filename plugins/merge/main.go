package merge

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/common/hexutil"

	ctypes "github.com/openrelayxyz/cardinal-types"
	"github.com/openrelayxyz/cardinal-types/metrics"

	"github.com/openrelayxyz/xplugeth"
	"github.com/openrelayxyz/xplugeth/types"
	"github.com/openrelayxyz/xplugeth/hooks/initialize"
)

// by importing the cardinal plugin below create an import chain which enables us to only require importing this plugin into geth
import (
	"github.com/openrelayxyz/xplugeth/plugins/producer"
)

type mergePlugin struct{}
type cardinalHook struct{}

var (
	postMerge       bool
	backend         types.Backend
	gethWeightGauge = metrics.NewMajorGauge("/geth/weight")
	stack           node.Node
	chainid         int64
)

type numLookup struct {
	Number hexutil.Big `json:"number"`
}

func init() {
	xplugeth.RegisterModule[mergePlugin]("mergePlugin")
}

func (*mergePlugin) InitializeNode(s *node.Node, b types.Backend) {
	stack = *s
	backend = b

	var hex hexutil.Uint64
	c := stack.Attach()
	c.Call(&hex, "eth_chainID")
	chainid = int64(hex)

	// chainid = b.ChainConfig().ChainID.Int64()
	if present := xplugeth.HasModule("cardinalProducerModule"); !present {
		panic("cardinal plugin not detected from merge plugin")
	}
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

func (*mergePlugin) CardinalAddBlockHook(number int64, hash, parent ctypes.Hash, weight *big.Int, updates map[string][]byte, deletes map[string]struct{}) {
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

var (
	_ initialize.Initializer = (*mergePlugin)(nil)
	
	_ producer.ExternalAddBlock = (*mergePlugin)(nil)
)
