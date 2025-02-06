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


	// go func() {
	// 	err := http.ListenAndServe(h.port, nil)
	// 	if err != nil {
	// 		log.Error("Error starting server", "err", err)
	// 	}
	// }()
	
	
}

func (h *healthCheckModule) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	log.Error("handle func")
	w.Header().Set("Content-Type", "application/json")

	// var response map[string]bool
	
	stat, err := h.getStatus()
	if err != nil {
		log.Error("error returned from call to getStatus, healthcheck unavailable", "err", err)
	}

	if stat {
		// response = map[string]bool{"ok": true}
		w.WriteHeader(200)
		w.Write([]byte(`{"ok": true}\n`))
	} else {
		// response = map[string]bool{"ok": false}
		w.WriteHeader(500)
		w.Write([]byte(`{"ok": false}\n`))
	}

	// json.NewEncoder(w).Encode(response)


	// hasWarning := false
	// if tm.shutdown {
	//   w.WriteHeader(500)
	//   w.Write([]byte(`{"ok": false}\n`))
	//   return
	// }
	// for _, hc := range tm.healthChecks {
	//   status := hc.Healthy()
	//   if status == rpc.Unavailable {
	// 	w.WriteHeader(500)
	// 	w.Write([]byte(`{"ok": false}\n`))
	// 	return
	//   }
	//   if status == rpc.Warning {
	// 	hasWarning = true
	//   }
	// }
	// if hasWarning {
	//   w.WriteHeader(429)
	//   w.Write([]byte(`{"ok": false}\n`))
	//   return
	// }
	// w.WriteHeader(200)
	// w.Write([]byte(`{"ok": true}\n`))
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
