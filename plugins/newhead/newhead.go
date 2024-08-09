package newhead

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	gethType "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/openrelayxyz/xplugeth"
	"github.com/openrelayxyz/xplugeth/shared"
	"github.com/openrelayxyz/xplugeth/types"
)

type newHeadModule struct{}

var (
	stack       node.Node
	controlData = make(map[uint64]map[string]interface{})
)

func init() {
	xplugeth.RegisterModule[newHeadModule]()
}

func (*newHeadModule) InitializeNode(s *node.Node, b types.Backend) {
	stack = *s
	log.Info("new head module initialized")

	var err error
	controlData, err = controlDataDecompress()
	if err != nil {
		log.Error("Failed to load control data", "error", err)
	}
}

func (*newHeadModule) Blockchain() {
	shared.Blockchain(stack)
}

func controlDataDecompress() (map[uint64]map[string]interface{}, error) {
	file, err := os.ReadFile("./test/core-control.json.gz")
	if err != nil {
		log.Error("cannot read file control.json.gz")
		return nil, err
	}
	r, err := gzip.NewReader(bytes.NewReader(file))
	if err != nil {
		return nil, err
	}
	defer r.Close()

	raw, err := io.ReadAll(r)
	if err == io.EOF || err == io.ErrUnexpectedEOF {
		return nil, err
	}

	var newheadObj map[uint64]map[string]interface{}
	json.Unmarshal(raw, &newheadObj)
	return newheadObj, nil
}

func (*newHeadModule) PluginNewHead(block *gethType.Block, hash common.Hash, logs []*gethType.Log, td *big.Int) {

	n, err := shared.GetBlockNumber(stack)
	if err != nil {
		log.Error("error returned from getnbr, test plugin", "err", err)
	}
	nbr, _ := hexutil.DecodeUint64(n)

	newHead := map[string]interface{}{
		"block": block,
		"hash":  hash.Bytes(),
		"logs":  logs,
		"td":    td,
	}

	expectedHead, exists := controlData[nbr]
	if !exists {
		log.Error("No expected data for block", "block", nbr)
		os.Exit(1)
		os.Remove("./test/testDataDir")
	}

	for k, v := range newHead {
		switch k {
		case "hash":
			expectedHash, ok := expectedHead[k]
			if !ok {
				continue
			}
			expectedBytes, err := hexutil.Decode(expectedHash.(string))
			if err != nil {
				log.Error("Failed to decode expected value", "block", nbr, "key", k, "error", err)
				continue
			}
			actualBytes := v.([]byte)
			if !bytes.Equal(actualBytes, expectedBytes) {
				log.Error("error mismatch in hash of", "block", nbr, "key", k, "actual", hexutil.Encode(actualBytes), "expected", hexutil.Encode(expectedBytes))
			}
		}
	}

}

// func (*newHeadModule) PluginNewSideBlock(block *gethType.Block, hash common.Hash, logs []*gethType.Log) {

// }
