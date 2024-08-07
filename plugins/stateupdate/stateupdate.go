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
	controlData  = make(map[uint64]map[string]interface{})
)

func init() {
	xplugeth.RegisterModule[stateUpdateModule]()
}

func (*stateUpdateModule) InitializeNode(s *node.Node, b types.Backend) {
	stack = *s
	log.Info("state update module initialized")
	var err error
	controlData, err = stateDataDecompress()
	if err != nil {
		log.Error("Failed to load control data", "error", err)
	}
}

func (*stateUpdateModule) Blockchain() {

	log.Error("inside of blockchain function")

	var chainCall bool
	client := stack.Attach()
	if err := client.Call(&chainCall, "admin_importChain", "./test/holesky-1-2000-chain.gz"); err != nil {
		log.Error("Error calling importChain from client, stack demo plugin", "err", err)
	}

	var blockCall string
	if err := client.Call(&blockCall, "eth_blockNumber"); err != nil {
		log.Error("Error calling blockNumber from client, stack demo plugin", "err", err)
	}

	blockNumber, err := hexutil.DecodeUint64(blockCall)
	if err != nil {
		log.Error("number decodeing error, stack demo plugin", "err", err)
		os.Exit(1)
	}

	if chainCall != true {
		log.Error("chain not imported", "chain", chainCall)
		os.Exit(1)
	}
	if blockNumber != 2000 {
		log.Error("blockNumber mismatch, chain not imported properly", "actual number", blockNumber)
		os.Exit(1)
	} else {
		os.RemoveAll("./test/testDataDir")
		os.Exit(0)
	}
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
	compareStateUpdates(stateUpdate, nbr)
}

func compareStateUpdates(stateUpdate map[string]interface{}, nbr uint64) {
	expectedUpdate, exists := controlData[nbr]
	if !exists {
		log.Error("No expected data for block", "block", nbr)
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
