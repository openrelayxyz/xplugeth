package example

import (
	"github.com/openrelayxyz/xplugeth"
	"github.com/openrelayxyz/xplugeth/types"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/log"
)

var (
	sessionStack *node.Node
	sessionBackend types.Backend
)

type exampleModule struct {}

func init() {
	xplugeth.RegisterModule[exampleModule]()
}

type ExampleConfig struct {
	FieldZero   bool `yaml:"fieldZero"`
	FieldOne    string `yaml:"fieldOne"`
}

var cfg *ExampleConfig


func (*exampleModule) InitializeNode(s *node.Node, b types.Backend) {
	log.Info("Example module initialized")
	
	var ok bool
	cfg, ok = xplugeth.GetConfig[ExampleConfig]("example")
	if !ok {
		cfg = &ExampleConfig{ FieldOne: "not set" }
		log.Warn("could not acqire config, all values set to default")
	}

	log.Info("example config values", "fieldZero", cfg.FieldZero, "fieldOne", cfg.FieldOne,)

}

func (*exampleModule) Shutdown() {
	log.Info("Byeee!")
}
