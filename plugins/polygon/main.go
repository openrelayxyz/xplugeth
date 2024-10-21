package polygon

import (
	"encoding/json"
	"fmt"
	"math/big"
	"regexp"

	"github.com/openrelayxyz/xplugeth"
	"github.com/openrelayxyz/xplugeth/hooks/initialize"
	"github.com/openrelayxyz/xplugeth/types"

	ctypes "github.com/openrelayxyz/cardinal-types"
	"github.com/openrelayxyz/cardinal-types/hexutil"
	
	core "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
)

// by importing the cardinal plugin below create an import chain which enables us to only require importing this plugin into geth
import (
	"github.com/openrelayxyz/xplugeth/plugins/producer"
)

type polygonPlugin struct{
	backend   types.Backend
	stack     node.Node
	chainid   int64
	client    *rpc.Client
}

func init() {
	xplugeth.RegisterModule[polygonPlugin]("polygonPlugin")
}

func (p *polygonPlugin) InitializeNode(s *node.Node, b types.Backend) {
	p.stack = *s
	p.backend = b
	p.client = s.Attach()

	var hex hexutil.Uint64
	p.client.Call(&hex, "eth_chainID")
	p.chainid = int64(hex)

	if present := xplugeth.HasModule("cardinalProducerModule"); !present {
		panic("cardinal plugin not detected from polygon plugin")
	}
	log.Info("initialized Cardinal EVM polygon plugin")
}

func (*polygonPlugin) UpdateStreamsSchema(schema map[string]string) {
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

func (p *polygonPlugin) CardinalAddBlockHook(number int64, hash, parent ctypes.Hash, weight *big.Int, updates map[string][]byte, deletes map[string]struct{}) {
	if p.client == nil {
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
		if err := p.client.Call(&borsnap, "bor_getSnapshot", hexutil.Uint64(number)); err != nil {
			log.Error("Error retrieving bor snapshot", "block", number)
		}
		updates[fmt.Sprintf("c/%x/b/%x/bs", uint64(p.chainid), hash.Bytes())] = borsnap
	}

	var receipt core.Receipt
	if err := p.client.Call(&receipt, "eth_getBorBlockReceipt", hash); err != nil {
		log.Debug("No bor receipt", "blockno", number, "hash", hash, "err", err)
		return
	}
	updates[fmt.Sprintf("c/%x/b/%x/br/%x", uint64(p.chainid), hash.Bytes(), receipt.TransactionIndex)] = receipt.Bloom.Bytes()
	for _, logRecord := range receipt.Logs {
		updates[fmt.Sprintf("c/%x/b/%x/bl/%x/%x", p.chainid, hash.Bytes(), receipt.TransactionIndex, logRecord.Index)], _ = rlp.EncodeToBytes(logRecord)
	}
}

var (
	_ initialize.Initializer = (*polygonPlugin)(nil)
	
	_ producer.ExternalAddBlock = (*polygonPlugin)(nil)
	_ producer.ExternalStreamSchema = (*polygonPlugin)(nil)
)
