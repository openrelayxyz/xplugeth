package utils

import (
	"github.com/openrelayxyz/xplugeth"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/node"
)

func GetChainID() (int64, bool) {
	s, ok := xplugeth.GetSingleton[*node.Node]()
	if !ok {
		return 0, false
	}
	var hex hexutil.Uint64
	client := s.Attach()
	client.Call(&hex, "eth_chainID")
	return int64(hex), true
}