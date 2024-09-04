package getAPIsDemo

import (
	"context"
	"net"
	"net/http"
	"encoding/json"
	"bytes"
	"time"
	"os"

	"github.com/openrelayxyz/xplugeth"
	"github.com/openrelayxyz/xplugeth/types"
	
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
)

type getAPIsDemo struct {
	stack   *node.Node
	backend types.Backend
}

func (*getAPIsDemo) InitializeNode(s *node.Node, b types.Backend) {
	log.Info("getAPIsDemo module initialized")
}

func init() {
	xplugeth.RegisterModule[getAPIsDemo]()
}

func (*getAPIsDemo) GetAPIs(stack *node.Node, backend types.Backend) []rpc.API {
	return []rpc.API{
		{
			Namespace: "xplugeth",
			Version:	 "1.0",
			Service:	 &getAPIsDemo{stack, backend},
			Public:		true,
		},
 	}
}

func (g *getAPIsDemo) Blockchain() {
	log.Error("inside of plugin side blockchain")

	// Note Jesse: Ok so this is the basic formula for testing this hook and injection. As you can see below we are utilizing the
	// nodeInterval and modifiedInterval to confirm that the RPC SetTrieFlushInterval has been called. 
	
	// I would like to see if you can integrate what we have here into our existing test. 
	
	// In this case I am calling the runTest method manually from the command line once I have started geth and imported the chain 
	// manually. Ultimately what I would like to do is to use runTest to kick off the chain import and all other testing logic with a go routine. 

	// Take a look through what we have here and lets plan to meet in the next couple of days. I can meet tommorow (8/4/24) but not unitl later 
	// in my day. let me know what works for you. 
}

var nodeInterval time.Duration
var modifiedInterval time.Duration 

func evaluateInitialDuration(duration time.Duration) int {
	d := int(duration.Seconds())
	return d
}

func (*getAPIsDemo) SetTrieFlushIntervalClone(duration time.Duration) time.Duration {
	nodeInterval = duration 

	if modifiedInterval > 0 {
		duration = modifiedInterval
	}

	return duration
}

func (*getAPIsDemo) SetTrieFlushInterval(ctx context.Context, interval string) error {
	newInterval, err := time.ParseDuration(interval)
	if err != nil {
		return err
	}
	modifiedInterval = newInterval

	return nil
}

func (*getAPIsDemo) RunTest(ctx context.Context) {

	callRPC()

	time.Sleep(1 * time.Second)
	
	if modifiedInterval <= nodeInterval {
		log.Error("setTrieFlush not functional")
		os.Exit(1)
	} else {
		log.Info("exit without error")
		os.Exit(0)
	}
}


var client = &http.Client{Transport: &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext,
	MaxIdleConnsPerHost:   16,
	MaxIdleConns:          16,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
}}

type Call struct {
	Version string            `json:"jsonrpc"`
	ID      json.RawMessage   `json:"id"`
	Method  string            `json:"method"`
	Params  []json.RawMessage `json:"params"`
}



func callRPC() {

	id, err := json.Marshal(1)
	if err != nil { 
		log.Error("json marshalling error, id, callRPC, test plugin", "err", err)
	}

	val, err := json.Marshal("1s")
	if err != nil { 
		log.Error("json marshalling error, params val, callRPC, test plugin", "err", err) 
	}


	call := &Call{
		Version: "2.0",
		ID : id,
		Method: "xplugeth_setTrieFlushInterval",
		Params: []json.RawMessage{val},
	  }

	backendURL := "http://127.0.0.1:8545"

	callBytes, _ := json.Marshal(call)

	request, _ := http.NewRequestWithContext(context.Background(), "POST", backendURL, bytes.NewReader(callBytes))
	request.Header.Add("Content-Type", "application/json")

	_, err = client.Do(request)

	if err != nil {
		log.Error("Error calling passive node from PreTrieCommit", "err", err)
	}
}
