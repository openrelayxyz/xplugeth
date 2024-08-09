package shared

import (
	"os"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
)

func GetBlockNumber(s node.Node) (string, error) {
	client := s.Attach()
	var num string
	if err := client.Call(&num, "eth_blockNumber"); err != nil {
		return "", err
	} else {
		return num, nil
	}
}

func Blockchain(s node.Node) {
	log.Error("inside of blockchain function")
	var chainCall bool
	client := s.Attach()
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
