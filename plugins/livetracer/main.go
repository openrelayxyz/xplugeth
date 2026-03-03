package livetracer

import (
	"context"
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/tracing"
	gtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/eth/tracers"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/openrelayxyz/xplugeth"
	"github.com/openrelayxyz/xplugeth/hooks/apis"
	"github.com/openrelayxyz/xplugeth/types"
)

var (
	events event.Feed
)

func init() {
	log.Info("xplugeth live block initialized")
	tracers.LiveDirectory.Register("xtracer", newXPlugethLiveTracer)
	xplugeth.RegisterModule[tracerResult]("liveBlockTracer")
}

type tracerResult struct{
	CallStack []CallStack
	Results   []CallStack
}

func newXPlugethLiveTracer(_ json.RawMessage) (*tracing.Hooks, error) {
	t := &tracerResult{}
	return &tracing.Hooks{
		OnTxStart:        t.OnTxStart,
		OnTxEnd:          t.OnTxEnd,
		OnEnter:          t.OnEnter,
		OnExit:           t.OnExit,
		OnBlockStart:     t.OnBlockStart,
		OnBlockEnd:       t.OnBlockEnd,
	}, nil
}

func (t *tracerResult) OnEnter(depth int, typ byte, from common.Address, to common.Address, input []byte, gas uint64, value *big.Int) {
	t.CallStack = append(t.CallStack, CallStack{
		Type:  vm.OpCode(typ).String(),
		From:  from,
		To:    to,
		Input: hexutil.Bytes(input),
		Gas:   hexutil.Uint64(gas),
		Calls: []CallStack{},
	})
}

func (t *tracerResult) OnExit(depth int, output []byte, gasUsed uint64, err error, reverted bool) {
	if len(t.CallStack) > 1 {
		returnCall := t.CallStack[len(t.CallStack)-1]
		returnCall.GasUsed = hexutil.Uint64(gasUsed)
		returnCall.Output = output
		t.CallStack[len(t.CallStack)-2].Calls = append(t.CallStack[len(t.CallStack)-2].Calls, returnCall)
		t.CallStack = t.CallStack[:len(t.CallStack)-1]
	}
}

func (t *tracerResult) OnTxStart(vm *tracing.VMContext, tx *gtypes.Transaction, from common.Address) {
	t.CallStack = []CallStack{}
}

func (t *tracerResult) OnTxEnd(receipt *gtypes.Receipt, err error) {
	if len(t.CallStack) > 0 {
		t.Results = append(t.CallStack)
	}
}

func (t *tracerResult) OnBlockStart(ev tracing.BlockEvent) {
	t.Results = []CallStack{}
}

func (t *tracerResult) OnBlockEnd(err error) {
	if len(t.Results) > 0 {
		events.Send(t.Results)
	}
}

type CallStack struct {
	Type    string         `json:"type"`
	From    common.Address   `json:"from"`
	To      common.Address   `json:"to"`
	Value   *big.Int       `json:"value,omitempty"`
	Gas     hexutil.Uint64 `json:"gas"`
	GasUsed hexutil.Uint64 `json:"gasUsed"`
	Input   hexutil.Bytes  `json:"input"`
	Output  hexutil.Bytes  `json:"output"`
	Time    string         `json:"time,omitempty"`
	Calls   []CallStack    `json:"calls,omitempty"`
	Results []CallStack    `json:"results,omitempty"`
	Error   string         `json:"error,omitempty"`
}

type tracerWSAPI struct {}

func (t *tracerWSAPI) TraceBlock(ctx context.Context) (<-chan []CallStack, error) {
	subch := make(chan []CallStack, 1000)
	rtrnch := make(chan []CallStack, 1000)
	go func() {
		log.Info("Subscription Block Tracer setup")
		sub := events.Subscribe(subch)
		for {
			select {
			case <-ctx.Done():
				sub.Unsubscribe()
				close(subch)
				close(rtrnch)
				return
			case t := <-subch:
				rtrnch <- t
			case <-sub.Err():
				sub.Unsubscribe()
				close(subch)
				close(rtrnch)
				return
			}
		}
	}()
	return rtrnch, nil
}

func (*tracerResult) GetAPIs(*node.Node, types.Backend) []rpc.API {
	log.Info("Registering xplugeth live block tracer APIs")
	return []rpc.API{
		{
			Namespace: "plugeth",
			Service:   &tracerWSAPI{},
		},
	}
}

var (
	_ apis.GetAPIs = (*tracerResult)(nil)
)
