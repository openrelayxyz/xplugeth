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

var SUcount int

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

	expectedData, err := stateDataDecompress()
	if err != nil {
		log.Error("Failed to load control data", "error", err)
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

	currentBlock := blockNumbers[SUcount-1]
	expected := expectedData[currentBlock]

	if expectedBlockRootStr, ok := expected["blockRoot"].(string); ok {
		expectedBlockRoot := common.HexToHash(expectedBlockRootStr).Bytes()
		if !bytes.Equal(blockRoot.Bytes(), expectedBlockRoot) {
			log.Error("Mismatch in blockRoot", "block", currentBlock, "expected", common.BytesToHash(expectedBlockRoot), "actual", blockRoot)
		}
	}

	if expectedParentRootStr, ok := expected["parentRoot"].(string); ok {
		expectedParentRoot := common.HexToHash(expectedParentRootStr).Bytes()
		if !bytes.Equal(parentRoot.Bytes(), expectedParentRoot) {
			log.Error("Mismatch in parentRoot", "block", currentBlock, "expected", common.BytesToHash(expectedParentRoot), "actual", parentRoot)
		}
	}

	if expectedDestructs, ok := expected["destructs"].(map[string]interface{}); ok {
		for addrStr := range expectedDestructs {
			addr := common.HexToHash(addrStr)
			if _, exists := destructs[addr]; !exists {
				log.Error("Missing destruct", "block", currentBlock, "address", addr)
			}
		}
	}

	if expectedAccounts, ok := expected["accounts"].(map[string]interface{}); ok {
		expectedAccountsCount := len(expectedAccounts)
		actualAccountsCount := len(accounts)

		if expectedAccountsCount != actualAccountsCount {
			log.Error("Mismatch in number of accounts", "block", currentBlock, "expected", expectedAccountsCount, "actual", actualAccountsCount)
		} else {
			log.Info("Account counts match", "block", currentBlock, "count", actualAccountsCount)
		}
	} else {
		log.Error("Expected accounts data not found or has incorrect type", "block", currentBlock)
	}

}

func init() {
	xplugeth.RegisterModule[stateUpdateModule]()
}
