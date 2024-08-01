package stateupdate

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"io"
	"os"
	"reflect"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/openrelayxyz/xplugeth"
	"github.com/openrelayxyz/xplugeth/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
)

type stateUpdateModule struct{}

var (
	stack        node.Node
	SUcount      int
	StateUpdates = make(map[uint64]map[string]interface{})
)

func init() {
	xplugeth.RegisterModule[stateUpdateModule]()
}

func (*stateUpdateModule) InitializeNode(s *node.Node, b types.Backend) {
	stack = *s
	log.Info("state update module initialized")
}

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

func getBlockNumber() (string, error) {
	client := stack.Attach()
	var num string
	if err := client.Call(&num, "eth_blockNumber"); err != nil {
		return "", err
	} else {
		return num, nil
	}
}

func (*stateUpdateModule) PluginStateUpdate(blockRoot, parentRoot common.Hash, destructs map[common.Hash]struct{}, accounts map[common.Hash][]byte, storage map[common.Hash]map[common.Hash][]byte, codeUpdates map[common.Hash][]byte) {
	SUcount += 1

	n, err := getBlockNumber()
	if err != nil {
		log.Error("error returned from getBlockNumber, test plugin", "err", err)
	}
	nbr, _ := hexutil.DecodeUint64(n)

	stateUpdate := map[string]interface{}{
		"blockRoot":  blockRoot.Bytes(),
		"parentRoot": parentRoot.Bytes(),
		"accounts":   accounts,
		"storages":   storage,
		"code":       codeUpdates,
	}
	StateUpdates[nbr] = stateUpdate
	compareStateUpdates()
}

func compareStateUpdates() {
	expectedData, err := stateDataDecompress()
	if err != nil {
		log.Error("Failed to load control data", "error", err)
	}
	if SUcount >= len(expectedData) {
		for blockNumber, stateUpdate := range StateUpdates {
			expectedUpdate, exists := expectedData[blockNumber]
			if !exists {
				log.Info("No expected data for block", "block", blockNumber)
				continue
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
						log.Error("Failed to decode expected value", "block", blockNumber, "key", k, "error", err)
						continue
					}
					actualBytes := v.([]byte)
					if !bytes.Equal(actualBytes, expectedBytes) {
						log.Error("Root mismatch", "block", blockNumber, "key", k, "actual", hexutil.Encode(actualBytes), "expected", hexutil.Encode(expectedBytes))
					}

				case "accounts":
					if expectedAccounts, exists := expectedUpdate[k]; exists {
						expectedCodeUpdatesMap, ok := expectedAccounts.(map[string]interface{})
						if !ok {
							log.Error("Invalid type for expected accounts", "block", blockNumber)
							continue
						}
						actualAccounts, ok := v.(map[common.Hash][]byte)
						if !ok {
							log.Error("Invalid type for accounts", "block", blockNumber)
							continue
						}
						for hash, actualValue := range actualAccounts {
							expectedValue, exists := expectedCodeUpdatesMap[hash.Hex()]
							if !exists {
								continue
							}
							expectedBytes, err := base64.StdEncoding.DecodeString(expectedValue.(string))
							if err != nil {
								log.Error("Failed to decode expected account value", "block", blockNumber, "account", hash.Hex(), "error", err)
								continue
							}
							if !bytes.Equal(actualValue, expectedBytes) {
								log.Error("Accounts mismatch",
									"block", blockNumber,
									"account", hash.Hex(),
									"actual", hexutil.Encode(actualValue),
									"expected", hexutil.Encode(expectedBytes))
							}
						}
					}
				case "storages":
					if expectedStorage, exists := expectedUpdate[k]; exists {
						expectedStorageMap, ok := expectedStorage.(map[string]interface{})
						if !ok {
							log.Error("Invalid type for expected storage", "block", blockNumber)
							continue
						}

						actualStorage, ok := v.(map[common.Hash]map[common.Hash][]byte)
						if !ok {
							log.Error("Invalid type for storage", "block", blockNumber)
							continue
						}

						for outerHash, innerMap := range actualStorage {
							expectedInner, exists := expectedStorageMap[outerHash.Hex()]
							if !exists {
								continue
							}

							expectedInnerMap, ok := expectedInner.(map[string]interface{})
							if !ok {
								log.Error("Invalid type for expected inner storage map", "block", blockNumber, "outerKey", outerHash.Hex())
								continue
							}

							for innerHash, actualValue := range innerMap {
								expectedValue, exists := expectedInnerMap[innerHash.Hex()]
								if !exists {
									continue
								}
								expectedStr, ok := expectedValue.(string)
								if !ok {
									log.Error("Invalid type for expected storage value",
										"block", blockNumber,
										"outerKey", outerHash.Hex(),
										"innerKey", innerHash.Hex(),
										"type", reflect.TypeOf(expectedValue))
									continue
								}

								expectedBytes, err := base64.StdEncoding.DecodeString(expectedStr)
								if err != nil {
									log.Error("Failed to decode expected storage value", "block", blockNumber, "outerKey", outerHash.Hex(), "innerKey", innerHash.Hex(), "error", err)
									continue
								}

								if !bytes.Equal(actualValue, expectedBytes) {
									log.Error("Storage mismatch", "block", blockNumber, "outerKey", outerHash.Hex(), "innerKey", innerHash.Hex(), "actual", hexutil.Encode(actualValue), "expected", hexutil.Encode(expectedBytes))
								}
							}
						}
					}
				case "code":
					if expectedCodeUpdates, exists := expectedUpdate[k]; exists {
						expectedCodeUpdatesMap, ok := expectedCodeUpdates.(map[string]interface{})
						if !ok {
							log.Error("Invalid type for expected codeupdates", "block", blockNumber)
							continue
						}
						actual, ok := v.(map[common.Hash][]byte)
						if !ok {
							log.Error("Invalid type for accounts", "block", blockNumber)
							continue
						}
						for hash, value := range actual {
							expectedValue, exists := expectedCodeUpdatesMap[hash.Hex()]
							if !exists {
								continue
							}
							expectedBytes, err := base64.StdEncoding.DecodeString(expectedValue.(string))
							if err != nil {
								log.Error("Failed to decode expected codeupdates value", "block", blockNumber, "code", hash.Hex(), "error", err)
								continue
							}
							if !bytes.Equal(value, expectedBytes) {
								log.Error("Code update mismatch",
									"block", blockNumber,
									"codeHash", hash.Hex(),
									"actual", hexutil.Encode(value),
									"expected", hexutil.Encode(expectedBytes))
							}
						}
					}
				}
			}
		}
	}
}
