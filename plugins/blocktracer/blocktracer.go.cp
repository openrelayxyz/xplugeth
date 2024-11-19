package main

import (
	"context"
	"math/big"
	"strings"
	"time"

	"github.com/openrelayxyz/plugeth-utils/core"
	"github.com/openrelayxyz/plugeth-utils/restricted"
	"github.com/openrelayxyz/plugeth-utils/restricted/hexutil"
)

var (
	pl      core.PluginLoader
	backend restricted.Backend
	log     core.Logger
	events  core.Feed
)

var (
	httpApiFlagName = "http.api"
	wsApiFlagName   = "ws.api"
)

func Initialize(ctx core.Context, loader core.PluginLoader, logger core.Logger) {
	log = logger
	pl = loader
	events = pl.GetFeed()
	v := ctx.String(httpApiFlagName)
	if v == "" {
		ctx.Set(wsApiFlagName, "eth,net,web3,plugeth")
	} else if !strings.Contains(v, "plugeth") {
		ctx.Set(wsApiFlagName, v+",plugeth")
	}
	log.Info("Loaded Block Tracer")
}

type TracerResult struct {
	CallStack []CallStack
	Results   []CallStack
}

type CallStack struct {
	Type    string         `json:"type"`
	From    core.Address   `json:"from"`
	To      core.Address   `json:"to"`
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

func (t *TracerResult) TraceBlock(ctx context.Context) (<-chan []CallStack, error) {
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

func GetLiveTracer(core.Hash, core.StateDB) core.BlockTracer {
	return &TracerResult{}
}

func (r *TracerResult) PreProcessBlock(hash core.Hash, number uint64, encoded []byte) {
	r.Results = []CallStack{}
}

func (r *TracerResult) PreProcessTransaction(tx core.Hash, block core.Hash, i int) {
}

func (r *TracerResult) BlockProcessingError(tx core.Hash, block core.Hash, err error) {
}

func (r *TracerResult) PostProcessTransaction(tx core.Hash, block core.Hash, i int, receipt []byte) {
}

func (r *TracerResult) PostProcessBlock(block core.Hash) {
	if len(r.Results) > 0 {
		events.Send(r.Results)
	}

}

func (r *TracerResult) CaptureStart(from core.Address, to core.Address, create bool, input []byte, gas uint64, value *big.Int) {
	r.CallStack = []CallStack{}
}
func (r *TracerResult) CaptureState(pc uint64, op core.OpCode, gas, cost uint64, scope core.ScopeContext, rData []byte, depth int, err error) {
}

func (r *TracerResult) CaptureFault(pc uint64, op core.OpCode, gas, cost uint64, scope core.ScopeContext, depth int, err error) {
}

func (r *TracerResult) CaptureEnd(output []byte, gasUsed uint64, t time.Duration, err error) {
	if len(r.CallStack) > 0 {
		r.Results = append(r.CallStack)
	}
}

func (r *TracerResult) CaptureEnter(typ core.OpCode, from core.Address, to core.Address, input []byte, gas uint64, value *big.Int) {
	r.CallStack = append(r.CallStack, CallStack{
		Type:  restricted.OpCode(typ).String(),
		From:  from,
		To:    to,
		Input: hexutil.Bytes(input),
		Gas:   hexutil.Uint64(gas),
		Calls: []CallStack{},
	})
}

func (r *TracerResult) CaptureExit(output []byte, gasUsed uint64, err error) {
	if len(r.CallStack) > 1 {
		returnCall := r.CallStack[len(r.CallStack)-1]
		returnCall.GasUsed = hexutil.Uint64(gasUsed)
		returnCall.Output = output
		r.CallStack[len(r.CallStack)-2].Calls = append(r.CallStack[len(r.CallStack)-2].Calls, returnCall)
		r.CallStack = r.CallStack[:len(r.CallStack)-1]
	}
}

func (r *TracerResult) Result() (interface{}, error) {
	return "", nil
}

func GetAPIs(node core.Node, backend core.Backend) []core.API {
	defer log.Info("APIs Initialized")
	return []core.API{
		{
			Namespace: "plugeth",
			Version:   "1.0",
			Service:   &TracerResult{},
			Public:    true,
		},
	}
}
