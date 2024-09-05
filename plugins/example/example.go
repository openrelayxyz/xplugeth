package example

import (
	"context"
	"time"
	"github.com/openrelayxyz/xplugeth"
	"github.com/openrelayxyz/xplugeth/types"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
)


type exampleModule struct {}

func (*exampleModule) InitializeNode(s *node.Node, b types.Backend) {
	log.Info("Example module initialized", "s", s, "b", b)
}

func (*exampleModule) Shutdown() {
	log.Info("Byeee!")
}

func (*exampleModule) GetAPIs(*node.Node, types.Backend) []rpc.API {
	log.Info("Registering APIs")
	return []rpc.API{
		{
			Namespace: "plugeth",
			Service:   &exampleService{},
		},
	}
}

func init() {
	xplugeth.RegisterModule[exampleModule]()
}

type exampleService struct{}

func (es *exampleService) Hello() string {
	return "Hello world!"
}

func (es *exampleService) Ticker(ctx context.Context) (<-chan int, error) {
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