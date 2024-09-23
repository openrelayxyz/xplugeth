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
	FieldZero   string `yaml:"server"`
	FieldOne    int    `yaml:"port"`
}

var cfg ExampleConfig



func (*exampleModule) InitializeNode(s *node.Node, b types.Backend) {
	log.Info("Example module initialized")
	
	cfg, ok := xplugeth.GetConfig[ExampleConfig]("example"); !ok {
		log.Warn("could not acqire config")
	}

}


func (*exampleModule) Shutdown() {
	log.Info("Byeee!")
}
