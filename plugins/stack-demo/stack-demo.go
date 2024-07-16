package example

import (
	"os"
	"time"

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

// Jesse Note: In the case below we are aquiring a client by attaching to the node object. This gives us access to the various rpc apis in geth. Here we are
// consuming the chain database and then checking for the lastest block number.


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

	// Jesse Note: the condition below is where we will ultimately be surveying our hooks and injections. At his point we are just checking to make sure that
	// the chain was imported. What I would like you to do is to add another condition in which we confirm that the block number is what we expect and that 
	// the block has the number of transactions that we expect. In order to accomplish this you will need to use the eth_getBlockByNumber method as well as
	// do some type conversions. For example the blockCall string above can be converted to a uint64 using the DecodeUint64 function in /common/hexutil/DecodeUint64.
	// a block can be unmarshalled using tools in /core/types which should be somewhat familar to you. 

	if chainCall != true {
		log.Error("chain not imported", "chain", chainCall)
		os.Exit(1)
	} else {
		log.Info("test successful")
		os.Exit(0)
	}
}

