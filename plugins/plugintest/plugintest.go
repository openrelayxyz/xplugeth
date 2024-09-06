package plugintest

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"math/big"
	"net"
	"net/http"
	"os"
	"reflect"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	gethType "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/openrelayxyz/xplugeth"
	"github.com/openrelayxyz/xplugeth/shared"
	"github.com/openrelayxyz/xplugeth/types"
)

type plugintest struct{}

var (
	stack            node.Node
	controlData      = make(map[uint64]map[string]interface{})
	stateUpdateData  = make(map[uint64]map[string]interface{})
	nodeInterval     time.Duration
	modifiedInterval time.Duration
	client           = &http.Client{Transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConnsPerHost:   16,
		MaxIdleConns:          16,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}}
)

func init() {
	xplugeth.RegisterModule[plugintest]()
}

func (p *plugintest) InitializeNode(s *node.Node, b types.Backend) {
	stack = *s
	log.Info("new head module initialized")

	var err error

	controlData, err = controlDataDecompress()
	if err != nil {
		log.Error("Failed to load control data", "error", err)
	}

	stateUpdateData, err = stateDataDecompress()
	if err != nil {
		log.Error("Failed to load control data", "error", err)
	}

	go func() {
		time.Sleep(5 * time.Second)
		p.RunTest(context.Background())
	}()
}

func (*plugintest) Blockchain() {
	shared.Blockchain(stack)
}

func (p *plugintest) SetTrieFlushIntervalClone(duration time.Duration) time.Duration {
	nodeInterval = duration

	if modifiedInterval > 0 {
		duration = modifiedInterval
	}

	return duration
}

func (p *plugintest) SetTrieFlushInterval(ctx context.Context, interval string) error {
	log.Error("running set trie flush")
	newInterval, err := time.ParseDuration(interval)
	if err != nil {
		return err
	}
	modifiedInterval = newInterval

	return nil
}

func (p *plugintest) GetAPIs(stack *node.Node, backend types.Backend) []rpc.API {
	return []rpc.API{
		{
			Namespace: "xplugeth",
			Version:   "1.0",
			Service:   p,
			Public:    true,
		},
	}
}

func (p *plugintest) RunTest(ctx context.Context) {
	err := p.SetTrieFlushInterval(ctx, "1s") // calling setTrieFlush here and from a request
	if err != nil {
		log.Error("Failed to set trie flush interval", "error", err)
		os.Exit(1)
	}

	time.Sleep(1 * time.Second)

	if modifiedInterval <= nodeInterval {
		log.Warn("values", "nodeInterval", nodeInterval, "modified", modifiedInterval)
		log.Error("setTrieFlush not functional")
		os.Exit(1)
	} else {
		log.Info("Trie flush interval test passed")
	}
}

// func (p *plugintest) AutomatedTest() {
// 	requestBody, err := json.Marshal(map[string]interface{}{
// 		"jsonrpc": "2.0",
// 		"method":  "xplugeth_runTest",
// 		"params":  []interface{}{},
// 		"id":      1,
// 	})
// 	if err != nil {
// 		log.Error("Failed to marshal request", "error", err)
// 		return
// 	}

// 	resp, err := client.Post("http://localhost:8545", "application/json", bytes.NewBuffer(requestBody))
// 	if err != nil {
// 		log.Error("Failed to send request", "error", err)
// 		return
// 	}
// 	defer resp.Body.Close()

// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		log.Error("Failed to read response", "error", err)
// 		return
// 	}

// 	log.Info("xplugeth_runTest response", "response", string(body))
// }

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

// func (*plugintest) Reorg(commonBlock *gethType.Block, oldChain, newChain gethType.Blocks) {

// }

// func (*plugintest) SetTrieFlushIntervalClone(flushInterval time.Duration) time.Duration {
// 	return flushInterval
// }

func (*plugintest) NewHead(block *gethType.Block, hash common.Hash, logs []*gethType.Log, td *big.Int) {
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
		log.Error("No expected data for block in NewHead", "block", nbr)
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

// func (*plugintest) NewSideBlock(block *gethType.Block, hash common.Hash, logs []*gethType.Log) {

// }

func stateDataDecompress() (map[uint64]map[string]interface{}, error) {
	file, err := os.ReadFile("./test/control.json.gz")
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

	var stateObject map[uint64]map[string]interface{}
	json.Unmarshal(raw, &stateObject)
	return stateObject, nil
}
func (*plugintest) StateUpdate(blockRoot, parentRoot common.Hash, destructs map[common.Hash]struct{}, accounts map[common.Hash][]byte, storage map[common.Hash]map[common.Hash][]byte, codeUpdates map[common.Hash][]byte) {
	n, err := shared.GetBlockNumber(stack)
	if err != nil {
		log.Error("error returned from getnbr, test plugin", "err", err)
	}
	nbr, _ := hexutil.DecodeUint64(n)

	stateUpdate := map[string]interface{}{
		"blockRoot":  blockRoot.Bytes(),
		"parentRoot": parentRoot.Bytes(),
		"accounts":   accounts,
		"storages":   storage,
		"code":       codeUpdates,
	}
	expectedUpdate, exists := stateUpdateData[nbr]
	if !exists {
		log.Error("No expected data for block in stateUpdate", "block", nbr)
		os.Exit(1)
	}
	for k, v := range stateUpdate {
		switch k {
		case "blockRoot", "parentRoot":
			expectedValue, exists := expectedUpdate[k]
			if !exists {
				continue
			}
			expectedBytes, err := hexutil.Decode(expectedValue.(string))
			if err != nil {
				log.Error("Failed to decode expected value", "block", nbr, "key", k, "error", err)
				continue
			}
			actualBytes := v.([]byte)
			if !bytes.Equal(actualBytes, expectedBytes) {
				log.Error("Root mismatch", "block", nbr, "key", k, "actual", hexutil.Encode(actualBytes), "expected", hexutil.Encode(expectedBytes))
			}
		case "accounts":
			if expectedAccounts, exists := expectedUpdate[k]; exists {
				expectedCodeUpdatesMap, ok := expectedAccounts.(map[string]interface{})
				if !ok {
					log.Error("Invalid type for expected accounts", "block", nbr)
					continue
				}
				actualAccounts, ok := v.(map[common.Hash][]byte)
				if !ok {
					log.Error("Invalid type for accounts", "block", nbr)
					continue
				}
				for hash, actualValue := range actualAccounts {
					expectedValue, exists := expectedCodeUpdatesMap[hash.Hex()]
					if !exists {
						continue
					}
					expectedBytes, err := base64.StdEncoding.DecodeString(expectedValue.(string))
					if err != nil {
						log.Error("Failed to decode expected account value", "block", nbr, "account", hash.Hex(), "error", err)
						continue
					}
					if !bytes.Equal(actualValue, expectedBytes) {
						log.Error("Accounts mismatch", "block", nbr, "account", hash.Hex(), "actual", hexutil.Encode(actualValue), "expected", hexutil.Encode(expectedBytes))
					}
				}
			}
		case "storages":
			if expectedStorage, exists := expectedUpdate[k]; exists {
				expectedStorageMap, ok := expectedStorage.(map[string]interface{})
				if !ok {
					log.Error("Invalid type for expected storage", "block", nbr)
					continue
				}

				actualStorage, ok := v.(map[common.Hash]map[common.Hash][]byte)
				if !ok {
					log.Error("Invalid type for actual storage", "block", nbr)
					continue
				}

				for outerHash, innerMap := range actualStorage {
					expectedInner, exists := expectedStorageMap[outerHash.Hex()]
					if !exists {
						continue
					}
					expectedInnerMap, ok := expectedInner.(map[string]interface{})
					if !ok {
						log.Error("Invalid type for expected inner storage map", "block", nbr, "outerKey", outerHash.Hex())
						continue
					}

					for innerHash, actualValue := range innerMap {
						expectedValue, exists := expectedInnerMap[innerHash.Hex()]
						if !exists {
							continue
						}

						expectedStr, ok := expectedValue.(string)
						if !ok {
							if expectedValue == nil {
								continue
							}
							log.Error("Invalid type for expected storage value",
								"block", nbr,
								"outerKey", outerHash.Hex(),
								"innerKey", innerHash.Hex(),
								"type", reflect.TypeOf(expectedValue))
							continue
						}

						expectedBytes, err := base64.StdEncoding.DecodeString(expectedStr)
						if err != nil {
							log.Error("Failed to decode expected storage value", "block", nbr, "outerKey", outerHash.Hex(), "innerKey", innerHash.Hex(), "error", err)
							continue
						}

						if !bytes.Equal(actualValue, expectedBytes) {
							log.Error("Storage mismatch", "block", nbr, "outerKey", outerHash.Hex(), "innerKey", innerHash.Hex(), "actual", hexutil.Encode(actualValue), "expected", hexutil.Encode(expectedBytes))
						}
					}
				}
			}
		case "code":
			if expectedCodeUpdates, exists := expectedUpdate[k]; exists {
				expectedCodeUpdatesMap, ok := expectedCodeUpdates.(map[string]interface{})
				if !ok {
					log.Error("Invalid type for expected codeupdates", "block", nbr)
					continue
				}
				actual, ok := v.(map[common.Hash][]byte)
				if !ok {
					log.Error("Invalid type for accounts", "block", nbr)
					continue
				}
				for hash, value := range actual {
					expectedValue, exists := expectedCodeUpdatesMap[hash.Hex()]
					if !exists {
						continue
					}
					expectedBytes, err := base64.StdEncoding.DecodeString(expectedValue.(string))
					if err != nil {
						log.Error("Failed to decode expected codeupdates value", "block", nbr, "code", hash.Hex(), "error", err)
						continue
					}
					if !bytes.Equal(value, expectedBytes) {
						log.Error("Code update mismatch", "block", nbr, "codeHash", hash.Hex(), "actual", hexutil.Encode(value), "expected", hexutil.Encode(expectedBytes))
					}
				}
			}
		case "destructs":
			expectedDestructs, exists := expectedUpdate[k]
			if !exists {
				log.Error("missing expected destructs", "block", nbr)
				continue
			}
			expectedDestructsMap, ok := expectedDestructs.(map[string]interface{})
			if !ok {
				log.Error("invalid type for expected destructs", "block", nbr)
				continue
			}
			actualDestructs, ok := v.(map[common.Hash]struct{})
			if !ok {
				log.Error("invalid type for actual destructs", "block", nbr)
				continue
			}
			for hash := range actualDestructs {
				if _, exists := expectedDestructsMap[hash.Hex()]; !exists {
					log.Error("Unexpected destruct", "block", nbr, "address", hash.Hex())
				}
			}
			for hash := range expectedDestructsMap {
				addressHash := common.HexToHash(hash)
				if _, exists := actualDestructs[addressHash]; !exists {
					log.Error("Missing destruct", "block", nbr, "address", hash)
				}
			}
		}
	}
}
