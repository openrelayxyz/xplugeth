package extensions

import (
	"errors"

	"github.com/openrelayxyz/xplugeth"
	"github.com/openrelayxyz/xplugeth/types"

	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/event"
)

type GetTxEventSubBackend interface {
	SubscribeNewTxsEvent(chan<- core.NewTxsEvent) event.Subscription
}

func SubscribeNewTxsEvent(ch chan<- core.NewTxsEvent) (event.Subscription, error) {
	if backend, ok := xplugeth.GetSingleton[types.Backend](); ok {
		if txsub, ok := backend.(GetTxEventSubBackend); ok {
			return txsub.SubscribeNewTxsEvent(ch), nil
		} else {
			return nil, errors.New("Unable to retun tx subscription, xplugeth")
		}
	} else {
		return nil, errors.New("Unable to get backend, xplugeth")
	}
}