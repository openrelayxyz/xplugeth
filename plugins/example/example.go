package example

import (
	"context"
	"time"
	
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/openrelayxyz/xplugeth"
	"github.com/openrelayxyz/xplugeth/hooks/apis"
	"github.com/openrelayxyz/xplugeth/hooks/initialize"
	"github.com/openrelayxyz/xplugeth/types"
)


type exampleModule struct {}

func init() {
	xplugeth.RegisterModule[exampleModule]("example")
}

type ExampleConfig struct {
	FieldZero   bool `yaml:"fieldZero"`
	FieldOne    string `yaml:"fieldOne"`
}

var cfg *ExampleConfig


func (*exampleModule) InitializeNode(*node.Node, types.Backend) {
	log.Info("Example module initialized")
	
	var ok bool
	cfg, ok = xplugeth.GetConfig[ExampleConfig]("example")
	if !ok {
		cfg = &ExampleConfig{ FieldOne: "not set" }
		log.Warn("did not acqire config, example plugin, all values set to default")
	}

	log.Info("example config values", "fieldZero", cfg.FieldZero, "fieldOne", cfg.FieldOne,)

}

func (*exampleModule) Shutdown() {
	log.Info("Byeee!")
}

func (*exampleModule) GetAPIs(*node.Node, types.Backend) []rpc.API {
	log.Info("Registering plugin APIs")
	return []rpc.API{
		{
			Namespace: "plugeth",
			Service:   &exampleAPIService{},
		},
	}
}

type exampleAPIService struct{}

func (es *exampleAPIService) Hello() string {
	return "Hello world!"
}

func (es *exampleAPIService) Ticker(ctx context.Context) (<-chan int, error) {
	ch := make(chan int)
	go func() {
		ticker := time.NewTicker(time.Second)
		counter := 0
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				ch <- counter
				counter++
			}
		}
	}()
	return ch, nil
}

var (
	_ apis.GetAPIs = (*exampleModule)(nil)
	_ initialize.Initializer = (*exampleModule)(nil)
	_ initialize.Shutdown = (*exampleModule)(nil)
)
