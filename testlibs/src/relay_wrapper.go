package testlibs

import (
	utilities "github.com/waku-org/waku-go-bindings/testlibs/utilities"
	"go.uber.org/zap"
)

func (wrapper *WakuNodeWrapper) Relay_Subscribe(pubsubTopic string) error {
	utilities.Debug("Attempting to subscribe to relay topic", zap.String("topic", pubsubTopic))

	if err := utilities.CheckWakuNodeNull(nil, wrapper.WakuNode); err != nil {
		utilities.Error("Cannot subscribe; node is nil", zap.Error(err))
		return err
	}

	err := wrapper.WakuNode.RelaySubscribe(pubsubTopic)
	if err != nil {
		utilities.Error("Failed to subscribe to relay topic", zap.Error(err))
		return err
	}

	utilities.Debug("Successfully subscribed to relay topic", zap.String("topic", pubsubTopic))
	return nil
}

func (wrapper *WakuNodeWrapper) Relay_Unsubscribe(pubsubTopic string) error {
	utilities.Debug("Attempting to unsubscribe from relay topic", zap.String("topic", pubsubTopic))

	if err := utilities.CheckWakuNodeNull(nil, wrapper.WakuNode); err != nil {
		utilities.Error("Cannot unsubscribe; node is nil", zap.Error(err))
		return err
	}

	err := wrapper.WakuNode.RelayUnsubscribe(pubsubTopic)
	if err != nil {
		utilities.Error("Failed to unsubscribe from relay topic", zap.Error(err))
		return err
	}

	utilities.Debug("Successfully unsubscribed from relay topic", zap.String("topic", pubsubTopic))
	return nil
}
