package blocktracer

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
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/openrelayxyz/xplugeth"
	"github.com/openrelayxyz/xplugeth/hooks/apis"
	"github.com/openrelayxyz/xplugeth/types"
)

var (
	events event.Feed
	block string
)

func init() {
	log.Error("inside xplugeth side live tracer")
	tracers.LiveDirectory.Register("xTracer", newXplugethTracer)
	xplugeth.RegisterModule[tracerResult]("liveBlockTracer")
}

// noop is a no-op live tracer. It's there to
// catch changes in the tracing interface, as well as
// for testing live tracing performance. Can be removed
// as soon as we have a real live tracet.
type tracerResult struct{
	CallStack []CallStack
	Results   []CallStack
}

func newXplugethTracer(_ json.RawMessage) (*tracing.Hooks, error) {
	log.Error("inside of xplugeth tracer")
	t := &tracerResult{}
	return &tracing.Hooks{
		OnTxStart:        t.OnTxStart,
		OnTxEnd:          t.OnTxEnd,
		OnEnter:          t.OnEnter,
		OnExit:           t.OnExit,
		OnOpcode:         t.OnOpcode,
		OnFault:          t.OnFault,
		OnGasChange:      t.OnGasChange,
		OnBlockchainInit: t.OnBlockchainInit,
		OnBlockStart:     t.OnBlockStart,
		OnBlockEnd:       t.OnBlockEnd,
		OnSkippedBlock:   t.OnSkippedBlock,
		OnGenesisBlock:   t.OnGenesisBlock,
		OnBalanceChange:  t.OnBalanceChange,
		OnNonceChange:    t.OnNonceChange,
		OnCodeChange:     t.OnCodeChange,
		OnStorageChange:  t.OnStorageChange,
		OnLog:            t.OnLog,
	}, nil
}

func (t *tracerResult) OnOpcode(pc uint64, op byte, gas, cost uint64, scope tracing.OpContext, rData []byte, depth int, err error) {
}

func (t *tracerResult) OnFault(pc uint64, op byte, gas, cost uint64, _ tracing.OpContext, depth int, err error) {
}

func (t *tracerResult) OnEnter(depth int, typ byte, from common.Address, to common.Address, input []byte, gas uint64, value *big.Int) {
	t.CallStack = append(t.CallStack, CallStack{
		Block: block, 
		Type:  vm.OpCode(typ).String(),
		From:  from,
		To:    to,
		Input: hexutil.Bytes(input),
		Gas:   hexutil.Uint64(gas),
		Calls: []CallStack{},
	})
	// log.Error("ENTER", "opcode", vm.OpCode(typ).String())
}

func (t *tracerResult) OnExit(depth int, output []byte, gasUsed uint64, err error, reverted bool) {
	if len(t.CallStack) > 1 {
		returnCall := t.CallStack[len(t.CallStack)-1]
		returnCall.GasUsed = hexutil.Uint64(gasUsed)
		returnCall.Output = output
		t.CallStack[len(t.CallStack)-2].Calls = append(t.CallStack[len(t.CallStack)-2].Calls, returnCall)
		t.CallStack = t.CallStack[:len(t.CallStack)-1]
	}
	// log.Error("EXIT", "exit_len", len(t.CallStack))
}

func (t *tracerResult) OnTxStart(vm *tracing.VMContext, tx *gtypes.Transaction, from common.Address) {
	// log.Error("inside of on tx start")
	t.CallStack = []CallStack{}
}

func (t *tracerResult) OnTxEnd(receipt *gtypes.Receipt, err error) {
	// log.Error("inside of on block end", "len", len(t.Results))
	if len(t.CallStack) > 0 {
		t.Results = append(t.CallStack)
	}
}

func (t *tracerResult) OnBlockStart(ev tracing.BlockEvent) {
	// log.Error("inside of on block start", "block", ev.Block.Number().String())
	block = ev.Block.Number().String()
	t.Results = []CallStack{}
}

func (t *tracerResult) OnBlockEnd(err error) {
	// log.Error("inside of on block end", "len", len(t.Results))
	if len(t.Results) > 0 {
		events.Send(t.Results)
	}
}

func (t *tracerResult) OnSkippedBlock(ev tracing.BlockEvent) {}

func (t *tracerResult) OnBlockchainInit(chainConfig *params.ChainConfig) {
}

func (t *tracerResult) OnGenesisBlock(b *gtypes.Block, alloc gtypes.GenesisAlloc) {
}

func (t *tracerResult) OnBalanceChange(a common.Address, prev, new *big.Int, reason tracing.BalanceChangeReason) {
}

func (t *tracerResult) OnNonceChange(a common.Address, prev, new uint64) {
}

func (t *tracerResult) OnCodeChange(a common.Address, prevCodeHash common.Hash, prev []byte, codeHash common.Hash, code []byte) {
}

func (t *tracerResult) OnStorageChange(a common.Address, k, prev, new common.Hash) {
}

func (t *tracerResult) OnLog(l *gtypes.Log) {

}

func (t *tracerResult) OnGasChange(old, new uint64, reason tracing.GasChangeReason) {
}

type CallStack struct {
	Block   string         `json."block"`
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

func (t *tracerResult) TraceBlock(ctx context.Context) (<-chan []CallStack, error) {
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
	log.Info("Registering live block tracer APIs")
	return []rpc.API{
		{
			Namespace: "plugeth",
			Service:   &tracerResult{},
		},
	}
}

var (
	_ apis.GetAPIs = (*tracerResult)(nil)
)
