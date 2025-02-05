package beaconagent

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/log"
	"github.com/openrelayxyz/cardinal-streams/transports"
	types "github.com/openrelayxyz/cardinal-types"
	"github.com/openrelayxyz/xplugeth"
)

type broker struct {
	URL          string `yaml:"url"`
	DefaultTopic string `yaml:"default.topic"`
	DataTopic    string `yaml:"data.topic"`
	Rollback     int64  `yaml:"rollback"`
}

type statsdOpts struct {
	Address  string `yaml:"address"`
	Port     string `yaml:"port"`
	Prefix   string `yaml:"prefix"`
	Interval int64  `yaml:"interval.sec"`
	Minor    bool   `yaml:"include.minor"`
}

type cloudwatchOpts struct {
	Namespace   string            `yaml:"namespace"`
	Dimensions  map[string]string `yaml:"dimensions"`
	Interval    int64             `yaml:"interval.sec"`
	Percentiles []float64         `yaml:"percentiles"`
	Minor       bool              `yaml:"include.minor"`
}

type Config struct {
	Chainid            int               `yaml:"chainid"`
	ReorgThreshold     int64             `yaml:"reorg.threshold"`
	RollbackSeconds    int               `yaml:"rollback.seconds"`
	TerminalDifficulty string            `yaml:"terminal.difficulty"`
	BackendURL         string            `yaml:"backend.url"`
	Whitelist          map[uint64]string `yaml:"whitelist"`
	LogLevel           int               `yaml:"log.level"`
	Brokers            []broker          `yaml:"brokers"`
	Statsd             *statsdOpts       `yaml:"statsd"`
	CloudWatch         *cloudwatchOpts   `yaml:"cloudwatch"`
	brokers            []transports.BrokerParams
	whitelist          map[uint64]types.Hash
	terminalDifficulty *big.Int
}

func LoadConfig() (*Config, error) {
	var ok bool
	cfg, ok := xplugeth.GetConfig[Config]("beaconagent")
	if !ok {
		log.Warn("config not found")
		cfg = &Config{
			Chainid:            1,
			ReorgThreshold:     128,
			RollbackSeconds:    3600,
			TerminalDifficulty: "20000000000000",
			LogLevel:           2,
		}
	}
	cfg.terminalDifficulty, _ = new(big.Int).SetString(cfg.TerminalDifficulty, 10)
	cfg.whitelist = make(map[uint64]types.Hash)
	for k, v := range cfg.Whitelist {
		cfg.whitelist[k] = types.HexToHash(v)
	}
	if len(cfg.Brokers) == 0 {
		return nil, fmt.Errorf("Config must specify at least one broker")
	}
	cfg.brokers = make([]transports.BrokerParams, len(cfg.Brokers))
	for i, b := range cfg.Brokers {
		if b.DefaultTopic == "" {
			b.DefaultTopic = fmt.Sprintf("cardinal-%v", cfg.Chainid)
		}
		if b.DataTopic == "" {
			b.DataTopic = fmt.Sprintf("%v-data", b.DefaultTopic)
		}
		if b.Rollback == 0 {
			b.Rollback = 5000
		}
		cfg.brokers[i] = transports.BrokerParams{
			URL:          b.URL,
			DefaultTopic: b.DefaultTopic,
			Topics: []string{
				b.DefaultTopic,
				b.DataTopic,
			},
			Rollback: b.Rollback,
		}
	}
	return cfg, nil
}
