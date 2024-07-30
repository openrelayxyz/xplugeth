package stateupdate

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"sort"

	"github.com/openrelayxyz/xplugeth"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

type stateUpdateModule struct{}

var (
	SUcount      int
	stateUpdates []map[string]interface{}
)

func stateDataDecompress() (map[uint64]map[string]interface{}, error) {
	// replace with path to control data
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

	raw, _ := io.ReadAll(r)
	if err == io.EOF || err == io.ErrUnexpectedEOF {
		return nil, err
	}

	var stateObject map[uint64]map[string]interface{}
	json.Unmarshal(raw, &stateObject)
	return stateObject, nil
}

func (*stateUpdateModule) PluginStateUpdate(blockRoot, parentRoot common.Hash, destructs map[common.Hash]struct{}, accounts map[common.Hash][]byte, storage map[common.Hash]map[common.Hash][]byte, codeUpdates map[common.Hash][]byte) {
	SUcount += 1

	stateUpdate := make(map[string]interface{})
	stateUpdates = append(stateUpdates, stateUpdate)

	if len(blockRoot) > 0 {
		stateUpdate["blockRoot"] = blockRoot.Bytes()
	}
	if len(parentRoot) > 0 {
		stateUpdate["parentRoot"] = parentRoot.Bytes()
	}
	if len(destructs) > 0 {
		stateUpdate["destructs"] = destructs
	}
	if len(accounts) > 0 {
		stateUpdate["accounts"] = accounts
	}
	if len(storage) > 0 {
		stateUpdate["storage"] = storage
	}
	if len(codeUpdates) > 0 {
		stateUpdate["codeUpdates"] = codeUpdates
	}
}

func CompareStateUpdate() {
	expectedData, err := stateDataDecompress()
	if err != nil {
		log.Error("Failed to load expected data", "error", err)
	}

	blockNumbers := make([]uint64, 0, len(expectedData))
	for blockNumber := range expectedData {
		blockNumbers = append(blockNumbers, blockNumber)
	}
	sort.Slice(blockNumbers, func(i, j int) bool { return blockNumbers[i] < blockNumbers[j] })

	expectedDataSlice := make([]map[string]interface{}, len(blockNumbers))
	for i, blockNumber := range blockNumbers {
		expectedDataSlice[i] = expectedData[blockNumber]
	}

	for i, state := range stateUpdates {
		for k, v := range state {
			expected := expectedDataSlice[i]
			if k == "blockRoot" || k == "parentRoot" {
				blockRoot := v.([]byte)
				expectedBlockRoot := []byte(expected[k].(string))
				if !bytes.Equal(blockRoot, expectedBlockRoot) {
					log.Error("Mismatch in blockRoot", "index", i, "expected", expectedBlockRoot, "actual", blockRoot)
				} else {
					log.Info("block root match")
				}
			}
		}
	}
}

func init() {
	xplugeth.RegisterModule[stateUpdateModule]()
}
