//go:build patchset
package producer

import (
	cli "github.com/urfave/cli/v2"
	"github.com/ethereum/go-ethereum/common"
)

func stateTrieUpdatesByNumber(i int64) (map[common.Hash]struct{}, map[common.Hash][]byte, map[common.Hash]map[common.Hash][]byte, map[common.Hash][]byte, error) {
	return nil, nil, nil, nil, nil
}

func trieDump (ctx cli.Context, args []string) {}