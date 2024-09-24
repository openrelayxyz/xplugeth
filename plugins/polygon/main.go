package polygon

import (
	"encoding/json"
	"fmt"
	"math/big"
	"regexp"

	core "github.com/ethereum/go-ethereum/core/types"
	glog "github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/rlp"
	ctypes "github.com/openrelayxyz/cardinal-types"
	"github.com/openrelayxyz/cardinal-types/hexutil"
	"github.com/openrelayxyz/xplugeth"
	"github.com/openrelayxyz/xplugeth/types"
)

type polygonPlugin struct{}

func init() {
	xplugeth.RegisterModule[polygonPlugin]()
}

var (
	log       glog.Logger
	postMerge bool
	backend   types.Backend
	stack     node.Node
	chainid   int64
	client    types.Client
)

func (*polygonPlugin) InitializeNode(s *node.Node, b types.Backend) {
	stack = *s
	backend = b
	chainid = b.ChainConfig().ChainID.Int64()

	client = stack.Attach()
	log.Info("Cardinal EVM resetting log level")
}

func UpdateStreamsSchema(schema map[string]string) {
	acctRe := regexp.MustCompile("c/([0-9a-z]+)/a/")
	var cid string
	for k := range schema {
		if match := acctRe.FindStringSubmatch(k); match != nil {
			cid = match[1]
			break
		}
	}
	if cid == "" {
		panic("Error finding chainid in schema")
	}
	schema[fmt.Sprintf("c/%v/b/[0-9a-z]+/br/", cid)] = schema[fmt.Sprintf("c/%v/b/[0-9a-z]+/r/", cid)]
	schema[fmt.Sprintf("c/%v/b/[0-9a-z]+/bl/", cid)] = schema[fmt.Sprintf("c/%v/b/[0-9a-z]+/l/", cid)]
	schema[fmt.Sprintf("c/%v/b/[0-9a-z]+/bs", cid)] = schema[fmt.Sprintf("c/%v/b/[0-9a-z]+/h", cid)]
}

func CardinalAddBlockHook(number int64, hash, parent ctypes.Hash, weight *big.Int, updates map[string][]byte, deletes map[string]struct{}) {
	if client == nil {
		log.Warn("Failed to initialize RPC client, cannot process block")
		return
	}

	var sprint int64
	if number < 38189056 {
		sprint = 64
	} else {
		sprint = 16
	}

	if number%sprint == 0 {
		var borsnap json.RawMessage
		if err := client.Call(&borsnap, "bor_getSnapshot", hexutil.Uint64(number)); err != nil {
			log.Error("Error retrieving bor snapshot on block %v", number)
		}
		updates[fmt.Sprintf("c/%x/b/%x/bs", uint64(chainid), hash.Bytes())] = borsnap
	}

	var receipt core.Receipt
	if err := client.Call(&receipt, "eth_getBorBlockReceipt", hash); err != nil {
		log.Debug("No bor receipt", "blockno", number, "hash", hash, "err", err)
		return
	}
	updates[fmt.Sprintf("c/%x/b/%x/br/%x", uint64(chainid), hash.Bytes(), receipt.TransactionIndex)] = receipt.Bloom.Bytes()
	for _, logRecord := range receipt.Logs {
		updates[fmt.Sprintf("c/%x/b/%x/bl/%x/%x", chainid, hash.Bytes(), receipt.TransactionIndex, logRecord.Index)], _ = rlp.EncodeToBytes(logRecord)
	}
}
