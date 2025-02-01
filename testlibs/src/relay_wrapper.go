package testlibs

import (
	"context"
	"errors"
	"time"

	"github.com/waku-org/go-waku/waku/v2/protocol/pb"
	utilities "github.com/waku-org/waku-go-bindings/testlibs/utilities"
	"github.com/waku-org/waku-go-bindings/waku/common"
	"go.uber.org/zap"
)

func (wrapper *WakuNodeWrapper) Wrappers_RelaySubscribe(pubsubTopic string) error {
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

	// Ensure the subscription happened by checking the number of connected relay peers
	numRelayPeers, err := wrapper.Wrappers_GetNumConnectedRelayPeers(pubsubTopic)
	if err != nil || numRelayPeers == 0 {
		utilities.Error("Subscription verification failed: no connected relay peers found", zap.Error(err))
		return errors.New("subscription verification failed: no connected relay peers")
	}

	utilities.Debug("Successfully subscribed to relay topic", zap.String("topic", pubsubTopic))
	return nil
}

func (wrapper *WakuNodeWrapper) Wrappers_RelayUnsubscribe(pubsubTopic string) error {
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

	// Ensure the unsubscription happened by verifying the relay peers count
	numRelayPeers, err := wrapper.Wrappers_GetNumConnectedRelayPeers(pubsubTopic)
	if err != nil {
		utilities.Error("Failed to verify unsubscription from relay topic", zap.Error(err))
		return err
	}
	if numRelayPeers > 0 {
		utilities.Error("Unsubscription verification failed: relay peers still connected", zap.Int("relayPeers", numRelayPeers))
		return errors.New("unsubscription verification failed: relay peers still connected")
	}

	utilities.Debug("Successfully unsubscribed from relay topic", zap.String("topic", pubsubTopic))
	return nil
}

func (wrapper *WakuNodeWrapper) Wrappers_RelayPublish(pubsubTopic string, message *pb.WakuMessage) (common.MessageHash, error) {
	if err := utilities.CheckWakuNodeNull(nil, wrapper.WakuNode); err != nil {
		utilities.Error("cannot publish message; node is nil", zap.Error(err))
		return "", err
	}

	// Handling context internally with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	utilities.Debug("Attempting to publish message via relay", zap.String("pubsubTopic", pubsubTopic))

	msgHash, err := wrapper.WakuNode.RelayPublish(ctx, message, pubsubTopic)
	if err != nil {
		utilities.Error("failed to publish message via relay", zap.Error(err))
		return "", err
	}

	utilities.Debug("Successfully published message via relay", zap.String("messageHash", msgHash.String()))
	return msgHash, nil
}
