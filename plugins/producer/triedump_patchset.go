//go:build patchset
package producer

import (
	"github.com/ethereum/go-ethereum/common"
)

func stateTrieUpdatesByNumber(i int64) (map[common.Hash]struct{}, map[common.Hash][]byte, map[common.Hash]map[common.Hash][]byte, map[common.Hash][]byte, error) {
	return nil, nil, nil, nil, nil
}

func trieDump (args []string) error {
	return nil
}