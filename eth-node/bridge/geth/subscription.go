package gethbridge

import (
	"github.com/ethereum/go-ethereum/event"

	"github.com/waku-org/waku-go-bindings/eth-node/types"
)

type gethSubscriptionWrapper struct {
	subscription event.Subscription
}

// NewGethSubscriptionWrapper returns an object that wraps Geth's Subscription in a types interface
func NewGethSubscriptionWrapper(subscription event.Subscription) types.Subscription {
	if subscription == nil {
		panic("subscription cannot be nil")
	}

	return &gethSubscriptionWrapper{
		subscription: subscription,
	}
}

func (w *gethSubscriptionWrapper) Err() <-chan error {
	return w.subscription.Err()
}

func (w *gethSubscriptionWrapper) Unsubscribe() {
	w.subscription.Unsubscribe()
}
