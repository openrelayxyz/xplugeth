package example

import (
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/openrelayxyz/xplugeth"
	"github.com/openrelayxyz/xplugeth/types"
)

var stack node.Node

type demoModule struct{}

func (*demoModule) InitializeNode(s *node.Node, b types.Backend) {
	stack = *s
	log.Info("stack demo module initialized")
}

func init() {
	xplugeth.RegisterModule[demoModule]()
}

func (*demoModule) Blockchain() {

	var chainCall bool
	client := stack.Attach()
	if err := client.Call(&chainCall, "admin_importChain", "./test/chain.gz"); err != nil {
		log.Error("Error calling importChain from client, stack demo plugin", "err", err)
	}

	time.Sleep(5 * time.Second)

	var blockCall string
	if err := client.Call(&blockCall, "eth_blockNumber"); err != nil {
		log.Error("Error calling blockNumber from client, stack demo plugin", "err", err)
	}

	var transactionCountCall string
	if err := client.Call(&transactionCountCall, "eth_getBlockTransactionCountByNumber", blockCall); err != nil {
		log.Error("Error fetching transaction count by block number", "err", err)
		os.Exit(1)
	}

	transactionCount, err := hexutil.DecodeUint64(transactionCountCall)
	if err != nil {
		log.Error("Error decoding transaction count", "err", err)
		os.Exit(1)
	}

	log.Info("transaction_count", "length", transactionCount)

	expectedBlocknumber, err := hexutil.DecodeUint64(blockCall)
	if err != nil || expectedBlocknumber <= 0 {
		log.Error("Error decoding block number from client, stack demo plugin", "err", err)
		os.Exit(1)
	}

	var block types.Block
	if err := client.Call(&block, "eth_getBlockByNumber", blockCall, true); err != nil {
		log.Error("Error fetching block by number from client, stack demo plugin", "err", err)
		os.Exit(1)
	}

	blockNumber, err := hexutil.DecodeUint64(block.Number)
	if err != nil {
		log.Error("Error decoding block number from block", "err", err)
	}

	if expectedBlocknumber != blockNumber {
		log.Error("Block number mismatch", "expected", expectedBlocknumber, "actual", blockNumber)
		os.Exit(1)
	} else {
		log.Info("Block number matches")
	}

	length := len(block.Transactions)
	log.Info("Length of transactions", "length", length)

	if chainCall != true {
		log.Error("chain not imported", "chain", chainCall)
		os.Exit(1)
	} else {
		log.Info("test successful")
		os.Exit(0)
	}
}
