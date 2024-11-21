package opcodeexample

import (
	"encoding/json"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/eth/tracers"
	"github.com/ethereum/go-ethereum/params"
)

func init() {
	tracers.DefaultDirectory.Register("opcounter", newOpcounter, false)
}

type opcounter struct {
	counts    map[string]int
	interrupt uint32
	reason    error
}

// var (
// 	chainConfig = &params.ChainConfig{}
// )

// newOpcounter returns a new opcode counting tracer.
func newOpcounter(ctx *tracers.Context, _ json.RawMessage, _ *params.ChainConfig) (*tracers.Tracer, error) {
	t := &opcounter{counts: make(map[string]int)}
	return &tracers.Tracer{
		Hooks: &tracing.Hooks{
			OnOpcode: t.onOpcode,
		},
		GetResult: t.getResult,
		Stop:      t.stop,
	}, nil
}

func (t *opcounter) onOpcode(pc uint64, op byte, gas, cost uint64, scope tracing.OpContext, rData []byte, depth int, err error) {
	// Skip if tracing was interrupted
	if atomic.LoadUint32(&t.interrupt) > 0 {
		return
	}
	name := vm.OpCode(op).String()
	if _, ok := t.counts[name]; !ok {
		t.counts[name] = 0
	}
	t.counts[name]++
}

func (t *opcounter) getResult() (json.RawMessage, error) {
	res, err := json.Marshal(t.counts)
	if err != nil {
		return nil, err
	}
	return res, t.reason
}

func (t *opcounter) stop(err error) {
	t.reason = err
	atomic.StoreUint32(&t.interrupt, 1)
}