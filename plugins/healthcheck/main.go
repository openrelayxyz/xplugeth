package example

import (
	"flag"	
	"time"
	"encoding/json"
	"net/http"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/openrelayxyz/xplugeth"
	"github.com/openrelayxyz/xplugeth/hooks/initialize"
	"github.com/openrelayxyz/xplugeth/types"
)

var (
	flags = *flag.NewFlagSet("healthcheck-plugin-flags", flag.ContinueOnError)
	hcTolerance = flags.Int("healthcheck.tolerance", 36, "tolerance for healthcheck in seconds, default is 36")
	hcPort = flags.String("healthcheck.port", "9999", "health check port, default is 9999")
)


type healthCheckModule struct {
	client	*rpc.Client
	tolerance int
	port string
}

func init() {
	xplugeth.RegisterFlags(flags)
	xplugeth.RegisterModule[healthCheckModule]("healthcheck")
}

func (h *healthCheckModule) InitializeNode(s *node.Node, b types.Backend) {
	h.client = s.Attach()

	h.tolerance = *hcTolerance
	if h.tolerance == 36 {
		log.Info("healthcheck tolerance set to default, 36 seconds")
	}

	h.port = ":"+*hcPort
	if *hcPort == "9999" {
		log.Info("healthcheck port set to default, 9999")
	}

	http.HandleFunc("/", h.handleHealthCheck)

	log.Info("healthcheck plugin initialized")

	go func() {
		err := http.ListenAndServe(h.port, nil)
		if err != nil {
			log.Error("Error starting server", "err", err)
		}
	}()
}

func (h *healthCheckModule) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	var response map[string]bool
	
	stat, err := h.getStatus()
	if err != nil {
		log.Error("error returned from call to getStatus, healthcheck unavailable", "err", err)
	}

	if stat {
		response = map[string]bool{"ok": true}
	} else {
		response = map[string]bool{"ok": false}
	}

	json.NewEncoder(w).Encode(response)
  }

  func (h *healthCheckModule) getStatus() (bool, error) {
	  
	var jsonBlock map[string]json.RawMessage
	h.client.Call(&jsonBlock, "eth_getBlockByNumber", "latest", false)

	raw, _ := jsonBlock["timestamp"]
	var timestamp string
	if err := json.Unmarshal(raw, &timestamp); err != nil {
		return false, err
	}

	timeInt, err := hexutil.DecodeUint64(timestamp)
	if err != nil {
		return false, err
	}

	unixTime := time.Unix(int64(timeInt), 0)
	t := time.Now()
	
	
	timeDiff := t.Sub(unixTime)

	if timeDiff <= (time.Duration(h.tolerance) * time.Second) {
		return true, nil
	} 

	return false, nil
}

var (
	_ initialize.Initializer = (*healthCheckModule)(nil)
)
