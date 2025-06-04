package agent

import (
	"net"
	_ "net/http/pprof"
	"net/http"
	"flag"
	"fmt"
	"strconv"
	"time"
	"os"
	"os/signal"
	"syscall"
	
	"github.com/openrelayxyz/cardinal-types/metrics"
	"github.com/openrelayxyz/cardinal-streams/v2/delivery"
	// "github.com/openrelayxyz/cardinal-beacon-mux/agent"

	"github.com/ethereum/go-ethereum/node"
	// "github.com/ethereum/go-ethereum/log"
	// "github.com/ethereum/go-ethereum/rpc"
	
	"github.com/savaki/cloudmetrics"
	"github.com/pubnub/go-metrics-statsd"
	log "github.com/inconshreveable/log15"

	"github.com/openrelayxyz/xplugeth"
	// "github.com/openrelayxyz/xplugeth/hooks/apis"
	"github.com/openrelayxyz/xplugeth/hooks/initialize"
	"github.com/openrelayxyz/xplugeth/types"
)

var (
	flags = *flag.NewFlagSet("beacon-agent", flag.ContinueOnError)
	debug = flag.Bool("debug", false, "Enable debug APIs")
	exitWhenSynced = flag.Bool("exitwhensynced", false, "Terminate when caught up with the network")
)

type beaconAgentModule struct {
}

func init() {
	xplugeth.RegisterFlags(flags)
	xplugeth.RegisterModule[beaconAgentModule]("beacon-agent")
}

func (*beaconAgentModule) InitializeNode(*node.Node, types.Backend) {
	log.Info("Beacon Agent module initialized")
}

func (*beaconAgentModule) Blockchain () {
	log.Error("blockchain plugin from within beacon agent")
	go agent()
}

func (*beaconAgentModule) Shutdown() {
	log.Error("Shutdown from within beacon agent plugin")
}

func agent() {
	// debug := flag.Bool("debug", false, "Enable debug APIs")
	// exitWhenSynced := flag.Bool("exitwhensynced", false, "Terminate when caught up with the network")

	// flag.CommandLine.Parse(os.Args[1:])
	// cfg, err := LoadConfig(flag.CommandLine.Args()[0])
	// if err != nil {
	// 	log.Error("Error parsing config", "err", err)
	// 	os.Exit(1)
	// }

	var ok bool
	rawCfg, ok := xplugeth.GetConfig[AgentConfig]("beacon-agent")
	if !ok {
		log.Warn("did not acqire config, beacon agent plugin")
	}
	cfg, err := LoadConfig(*rawCfg)
	if err != nil {
		log.Error("could not aquire config beacon agent plugin", "err", err)
	}

	var logLvl log.Lvl
	switch cfg.LogLevel {
	case 1:
		logLvl = log.LvlDebug
	case 2:
		logLvl = log.LvlInfo
	case 3:
		logLvl = log.LvlWarn
	case 4:
		logLvl = log.LvlError
	case 5:
		logLvl = log.LvlCrit
	default:
		logLvl = log.LvlInfo
	}
	log.Root().SetHandler(log.LvlFilterHandler(logLvl, log.Root().GetHandler()))

	if len(cfg.Brokers) == 0 {
		log.Error("No brokers specified")
		os.Exit(1)
	}
	if *debug {
		go func() {
			http.ListenAndServe("localhost:6060", nil)
		}()
	}
	sm, err := NewStreamManager(cfg.brokers, cfg.BackendURL, cfg.RollbackSeconds, cfg.terminalDifficulty, cfg.whitelist)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	errCh := sm.Start()
	select {
	case <-sm.Ready():
		log.Info("Synced to current")
	case err := <-errCh:
		log.Error("Sync error before reaching current.", "err", err.Error())
		os.Exit(1)

	}

	if *exitWhenSynced {
		log.Info("--exitwhensynced set: Exiting", "processed", sm.Processed())
		if sm.Processed() > 0 {
			os.Exit(0)
		}
		log.Warn("No blocks processed.")
		os.Exit(1)
	}

	metrics.Clear()
	delivery.Ready()

	if cfg.Statsd != nil && cfg.Statsd.Port != "" {
		addr := "127.0.0.1:" + cfg.Statsd.Port
		if cfg.Statsd.Address != "" {
			addr = fmt.Sprintf("%v:%v", cfg.Statsd.Address, cfg.Statsd.Port)
		}
		udpAddr, err := net.ResolveUDPAddr("udp", addr)
		if err != nil {
			log.Error("Invalid Address. Statsd will not be configured.", "error", err.Error())
		} else {
			interval := time.Duration(cfg.Statsd.Interval) * time.Second
			if cfg.Statsd.Interval == 0 {
				interval = time.Second
			}
			prefix := cfg.Statsd.Prefix
			if prefix == "" {
				prefix = "cardinal.evm"
			}
			go statsd.StatsD(
				metrics.MajorRegistry,
				interval,
				prefix,
				udpAddr,
			)
			if cfg.Statsd.Minor {
				go statsd.StatsD(
					metrics.MinorRegistry,
					interval,
					prefix,
					udpAddr,
				)
			}
		}
	}
	if cfg.CloudWatch != nil {
		namespace := cfg.CloudWatch.Namespace
		if namespace == "" {
			namespace = "Cardinal"
		}
		dimensions := []string{}
		for k, v := range cfg.CloudWatch.Dimensions {
			dimensions = append(dimensions, k, v)
		}
		if len(dimensions) == 0 {
			dimensions = append(dimensions, "chainid", strconv.Itoa(sm.ChainID()))
		}
		cwcfg := []func(*cloudmetrics.Publisher){
			cloudmetrics.Dimensions(dimensions...),
		}
		if cfg.CloudWatch.Interval > 0 {
			cwcfg = append(cwcfg, cloudmetrics.Interval(time.Duration(cfg.CloudWatch.Interval) * time.Second))
		}
		if len(cfg.CloudWatch.Percentiles) > 0 {
			cwcfg = append(cwcfg, cloudmetrics.Percentiles(cfg.CloudWatch.Percentiles))
		}
		go cloudmetrics.Publish(metrics.MajorRegistry,
			namespace,
			cwcfg...
		)
		if cfg.CloudWatch.Minor {
			go cloudmetrics.Publish(metrics.MinorRegistry,
				namespace,
				cwcfg...
			)
		}
	}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-sigs:
			log.Warn("Got shutdown signal. Shutting down.")
			sm.Close()
			time.Sleep(30)
			return
		case err := <-errCh:
			log.Warn("Error from Agent Stream. Restarting stream.", "err", err)
			sm.Close()
			sm, err = NewStreamManager(cfg.brokers, cfg.BackendURL, cfg.RollbackSeconds, cfg.terminalDifficulty, cfg.whitelist)
			if err != nil {
				log.Error(err.Error())
				os.Exit(1)
			}
			errCh = sm.Start()
		}
	}
}

var (
	_ initialize.Initializer = (*beaconAgentModule)(nil)
	_ initialize.Blockchain = (*beaconAgentModule)(nil)
	_ initialize.Shutdown = (*beaconAgentModule)(nil)
)