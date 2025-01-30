package utils

import (
	"fmt"
	"encoding/json"
	"errors"
	"math/big"

	"github.com/openrelayxyz/xplugeth"

	"github.com/ethereum/go-ethereum/common"
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
	client.Call(&hex, "eth_chainId")
	return int64(hex), true
}

func GetTd(hash common.Hash) (*big.Int, error) {
	result := new(big.Int)
	s, ok := xplugeth.GetSingleton[*node.Node]()
	if !ok {
		return nil, errors.New("failed to acqire stack singleton, GetTd")
	}
	var parentBlockJson map[string]json.RawMessage
	client := s.Attach()
	client.Call(&parentBlockJson, "eth_getBlockByHash", hash, false)
	raw, ok := parentBlockJson["totalDifficulty"]
	if !ok {
		chainid, ok := GetChainID()
		if !ok { panic(fmt.Sprintf("could not resolve chain id from within GetTd")) }
		switch chainid {
		case int64(1):
			result.SetString("58750003716598352816469", 10)
		case int64(11155111):
			result.SetString("17000018015853232", 10)
		default:
			result.SetString("1", 10)
		}
		return result, nil
	}
	var td string
	if err := json.Unmarshal(raw, &td); err != nil {
		return nil, err
	}
	if _, ok := result.SetString(td, 0); !ok {
		return nil, errors.New("convert total difficulty string to big int")
	} 
	return result, nil
}