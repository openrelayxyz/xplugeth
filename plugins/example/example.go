package example

import (
	"context"
	"errors"
	"flag"	
	"time"

	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/openrelayxyz/xplugeth"
	"github.com/openrelayxyz/xplugeth/hooks/apis"
	"github.com/openrelayxyz/xplugeth/hooks/initialize"
	"github.com/openrelayxyz/xplugeth/types"
)


var (
	flags = *flag.NewFlagSet("example-plugin", flag.ContinueOnError)
	exampleBoolFlag = flags.Bool("example.bool.flag", false, "example bool flag for xplugeth")
	exampleStringFlag = flags.String("example.string.flag", "", "example string flag for xplugeth")
)


type exampleModule struct {
}

func init() {
	xplugeth.RegisterFlags(flags)
	xplugeth.RegisterSubCommands(subCommands)
	xplugeth.RegisterModule[exampleModule]("example")
}

type ExampleConfig struct {
	FieldZero   bool `yaml:"fieldZero"`
	FieldOne    string `yaml:"fieldOne"`
}

var cfg *ExampleConfig


func (*exampleModule) InitializeNode(*node.Node, types.Backend, any) {
	log.Info("Example module initialized")

	if *exampleBoolFlag {
		log.Info("example bool flag set, example plugin")
	}
	if *exampleStringFlag != "" {
		log.Info("example string flag set, example plugin", "value", *exampleStringFlag)
	}
	
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

func (*exampleModule) GetAPIs(*node.Node, types.Backend, any) []rpc.API {
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
	subCommands map[string]func([]string)error = map[string]func([]string)error {
		"exampleSubComOne": func([]string) error {
				log.Info("you ran the FIRST subcommand")
				return nil
		},
		"exampleSubComTwo": func([]string) error {
				log.Info("you ran the SECOND subcommand")
				return nil
		},
		"exampleSubComThree": func([]string) error {
			return errors.New("the third subcommand returns this error")
		},
	}
)

var (
	_ apis.GetAPIs = (*exampleModule)(nil)
	_ initialize.Initializer = (*exampleModule)(nil)
	_ initialize.Shutdown = (*exampleModule)(nil)
)
