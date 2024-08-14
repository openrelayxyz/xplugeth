package coreinjection

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"io"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	gethType "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/openrelayxyz/xplugeth"
	"github.com/openrelayxyz/xplugeth/shared"
	"github.com/openrelayxyz/xplugeth/types"
)

type coreinjection struct{}

var (
	stack       node.Node
	controlData = make(map[uint64]map[string]interface{})
)

func init() {
	xplugeth.RegisterModule[coreinjection]()
}

func (*coreinjection) InitializeNode(s *node.Node, b types.Backend) {
	stack = *s
	log.Info("new head module initialized")

	var err error
	controlData, err = controlDataDecompress()
	if err != nil {
		log.Error("Failed to load control data", "error", err)
	}
}

func (*coreinjection) Blockchain() {
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

// func (*reorgModule) PluginReorg(commonBlock *gethType.Block, oldChain, newChain gethType.Blocks) {

// }

func (*coreinjection) PluginSetTrieFlushIntervalClone(flushInterval time.Duration) time.Duration {
	return flushInterval
}

func (*coreinjection) PluginNewHead(block *gethType.Block, hash common.Hash, logs []*gethType.Log, td *big.Int) {
	n, err := shared.GetBlockNumber(stack)
	if err != nil {
		log.Error("error returned from getnbr, test plugin", "err", err)
	}
	nbr, _ := hexutil.DecodeUint64(n)

	newHead := map[string]interface{}{
		"blockBytes": block,
		"hash":       hash.Bytes(),
		"logBytes":   logs,
		"totalDiff":  td.Bytes(),
	}

	expectedHead, exists := controlData[nbr]
	if !exists {
		log.Error("No expected data for block", "block", nbr)
		os.Exit(1)
		os.Remove("./test/testDataDir")
	}

	for k, v := range newHead {
		switch k {
		case "blockBytes":
			if nbr%10 == 0 {
				expectedBlockBytes, ok := expectedHead[k]
				if !ok {
					log.Error("Expected blockBytes is not of type string", "block", nbr)
					continue
				}
				expectedBytes, err := base64.StdEncoding.DecodeString(expectedBlockBytes.(string))
				if err != nil {
					log.Error("Failed to decode expected value", "block", nbr, "key", k, "error", err)
					continue
				}
				actualBytes, _ := rlp.EncodeToBytes(v)
				if !bytes.Equal(actualBytes, actualBytes) {
					log.Error("error mismatch", "block", nbr, "key", k, "actual", hexutil.Encode(actualBytes), "expected", hexutil.Encode(expectedBytes))
				}

			}
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
		case "logBytes":
			if nbr%10 == 0 {
				expectedLog, exists := expectedHead[k]
				if !exists {
					continue
				}
				expectedLogBytes, ok := expectedLog.([]interface{})
				if !ok {
					log.Error("logBytes not of type []interface{} for block", "block", nbr)
					continue
				}

				var decodedExpectedLogBytes [][]byte
				for _, p := range expectedLogBytes {
					logStr, ok := p.(string)
					if !ok {
						log.Error("logBytes entry is not a string in expected data", "block", nbr)
						continue
					}

					decodedBytes, err := base64.StdEncoding.DecodeString(logStr)
					if err != nil {
						log.Error("Failed to decode Base64 logBytes", "block", nbr, "error", err)
						continue
					}
					decodedExpectedLogBytes = append(decodedExpectedLogBytes, decodedBytes)
				}

				var actualLogBytes [][]byte
				for _, logItem := range v.([]*gethType.Log) {
					logBytes, err := rlp.EncodeToBytes(logItem)
					if err != nil {
						log.Error("Failed to encode log item", "block", nbr, "error", err)
						continue
					}
					actualLogBytes = append(actualLogBytes, logBytes)
				}

				if len(actualLogBytes) != len(decodedExpectedLogBytes) {
					log.Error("Mismatch in number of log bytes", "block", nbr, "actualCount", len(actualLogBytes), "expectedCount", len(decodedExpectedLogBytes))
					continue
				}

				for i, l := range actualLogBytes {
					if !bytes.Equal(l, decodedExpectedLogBytes[i]) {
						log.Error("Mismatch in logBytes", "block", nbr, "logIndex", i, "actual", base64.StdEncoding.EncodeToString(actualLogBytes[i]), "expected", base64.StdEncoding.EncodeToString(decodedExpectedLogBytes[i]))
					}
				}
			}

		case "totalDiff":
			expectedTd, ok := expectedHead[k].(string)
			if !ok {
				continue
			}
			expectedBigInt := new(big.Int)
			expectedBigInt.SetString(expectedTd, 10)
			expectedBytes := expectedBigInt.Bytes()
			actualBytes := v.([]byte)

			if !bytes.Equal(actualBytes, expectedBytes) {
				log.Error("error mismatch in total diff of", "block", nbr, "key", k, "actual", string(actualBytes), "expected", string(expectedBytes))
			}
		}
	}

}

// func (*newHeadModule) PluginNewSideBlock(block *gethType.Block, hash common.Hash, logs []*gethType.Log) {

// }
