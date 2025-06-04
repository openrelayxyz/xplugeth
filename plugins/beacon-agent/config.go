package agent

import (
	"fmt"
	"math/big"
	// "gopkg.in/yaml.v2"
	"github.com/openrelayxyz/cardinal-types"
	"github.com/openrelayxyz/cardinal-streams/v2/transports"
	// "io/ioutil"
)

type broker struct {
	URL string `yaml:"url"`
	DefaultTopic string `yaml:"default.topic"`
	DataTopic string `yaml:"data.topic"`
	Rollback int64 `yaml:"rollback"`
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

type AgentConfig struct {
	Chainid int `yaml:"chainid"`
	ReorgThreshold int64 `yaml:"reorg.threshold"`
	RollbackSeconds int `yaml:"rollback.seconds"`
	TerminalDifficulty string `yaml:"terminal.difficulty"`
	BackendURL string `yaml:"backend.url"`
	Whitelist map[uint64]string `yaml:"whitelist"`
	LogLevel int `yaml:"log.level"`
	Brokers []broker `yaml:"brokers"`
	Statsd *statsdOpts `yaml:"statsd"`
	CloudWatch *cloudwatchOpts `yaml:"cloudwatch"`
	brokers []transports.BrokerParams
	whitelist map[uint64]types.Hash
	terminalDifficulty *big.Int
}

func LoadConfig(cfg AgentConfig) (*AgentConfig, error) {
	// data, err := ioutil.ReadFile(fname)
	// if err != nil {
	// 	return nil, err
	// }
	// cfg := Config{}
	// if err := yaml.Unmarshal(data, &cfg); err != nil {
	// 	return nil, err
	// }
	if cfg.Chainid == 0 {
		cfg.Chainid = 1
	}
	if cfg.ReorgThreshold == 0 {
		cfg.ReorgThreshold = 128
	}
	if cfg.RollbackSeconds == 0 {
		cfg.RollbackSeconds = 3600
	}
	if cfg.TerminalDifficulty == "" {
		cfg.TerminalDifficulty = "20000000000000"
	}
	cfg.terminalDifficulty, _ = new(big.Int).SetString(cfg.TerminalDifficulty, 10)
	cfg.whitelist = make(map[uint64]types.Hash)
	for k, v := range cfg.Whitelist {
		cfg.whitelist[k] = types.HexToHash(v)
	}
	if cfg.LogLevel == 0 {
		cfg.LogLevel = 2
	}
	if len(cfg.Brokers) == 0 {
		return nil, fmt.Errorf("Config must specify at least one broker")
	}
	cfg.brokers = make([]transports.BrokerParams, len(cfg.Brokers))
	for i := range cfg.Brokers {
		if cfg.Brokers[i].DefaultTopic == "" {
			cfg.Brokers[i].DefaultTopic = fmt.Sprintf("cardinal-%v", cfg.Chainid)
		}
		if cfg.Brokers[i].DataTopic == "" {
			cfg.Brokers[i].DataTopic = fmt.Sprintf("%v-data", cfg.Brokers[i].DefaultTopic)
		}
		if cfg.Brokers[i].Rollback == 0 {
			cfg.Brokers[i].Rollback = 5000
		}
		cfg.brokers[i] = transports.BrokerParams{
			URL: cfg.Brokers[i].URL,
			DefaultTopic: cfg.Brokers[i].DefaultTopic,
			Topics: []string{
				cfg.Brokers[i].DefaultTopic,
				cfg.Brokers[i].DataTopic,
			},
			Rollback: cfg.Brokers[i].Rollback,
		}
	}
	return &cfg, nil
}
