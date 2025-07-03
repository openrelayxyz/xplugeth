package plugins

import (
	"context"
	"math/big"
	"errors"

	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/common"
)

type GetTxEventSubBackend interface {
	GetTd(ctx context.Context, hash common.Hash) *big.Int
}

func SubscribeNewTxsEvent(chan<- core.NewTxsEvent) event.Subscription, error {
	if backend, ok := xplugeth.GetSingleton[types.Backend](); ok {
		if txsub, ok := backend.(GetTxEventSubBackend); ok {
			return tdbackend.SubscribeNewTxsEvent(chain), nil
		} else {
			return nil, errors.New("Unable to retun tx subscription, xplugeth")
		}
	} else {
		return nil, errors.New("Unable to get backend, xplugeth")
	}
}