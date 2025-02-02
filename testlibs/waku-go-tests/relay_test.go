package waku_go_tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	testlibs "github.com/waku-org/waku-go-bindings/testlibs/src"
	utilities "github.com/waku-org/waku-go-bindings/testlibs/utilities"
	"go.uber.org/zap"
)

func TestRelaySubscribeToDefaultTopic(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	utilities.Debug("Starting test to verify relay subscription to the default pubsub topic")

	// Define the configuration with relay = true
	wakuConfig := *utilities.DefaultWakuConfig
	wakuConfig.Relay = true

	utilities.Debug("Creating a Waku node with relay enabled")
	node, err := testlibs.Wrappers_StartWakuNode(&wakuConfig, logger.Named("TestNode"))
	require.NoError(t, err)
	defer func() {
		utilities.Debug("Stopping and destroying the Waku node")
		node.Wrappers_StopAndDestroy()
	}()

	defaultPubsubTopic := utilities.DefaultPubsubTopic
	utilities.Debug("Default pubsub topic retrieved", zap.String("topic", defaultPubsubTopic))

	utilities.Debug("Fetching number of connected relay peers before subscription", zap.String("topic", defaultPubsubTopic))

	utilities.Debug("Attempting to subscribe to the default pubsub topic", zap.String("topic", defaultPubsubTopic))
	err = node.Wrappers_RelaySubscribe(defaultPubsubTopic)
	require.NoError(t, err)

	utilities.Debug("Fetching number of connected relay peers after subscription", zap.String("topic", defaultPubsubTopic))

	utilities.Debug("Test successfully verified subscription to the default pubsub topic", zap.String("topic", defaultPubsubTopic))
}

func TestRelayMessageTransmission(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	logger.Debug("Starting TestRelayMessageTransmission")

	// Configuration for sender node
	senderConfig := *utilities.DefaultWakuConfig
	senderConfig.Relay = true
	senderNode, err := testlibs.Wrappers_StartWakuNode(&senderConfig, logger.Named("SenderNode"))
	require.NoError(t, err)
	defer senderNode.Wrappers_StopAndDestroy()

	// Configuration for receiver node
	receiverConfig := *utilities.DefaultWakuConfig
	receiverConfig.Relay = true
	receiverNode, err := testlibs.Wrappers_StartWakuNode(&receiverConfig, logger.Named("ReceiverNode"))
	require.NoError(t, err)
	defer receiverNode.Wrappers_StopAndDestroy()

	logger.Debug("Connecting sender and receiver")
	err = senderNode.Wrappers_ConnectPeer(receiverNode)
	require.NoError(t, err)

	logger.Debug("Subscribing receiver to the default pubsub topic")
	defaultPubsubTopic := utilities.DefaultPubsubTopic
	err = receiverNode.Wrappers_RelaySubscribe(defaultPubsubTopic)
	require.NoError(t, err)

	logger.Debug("Creating and publishing message")
	message := senderNode.Wrappers_CreateMessage()
	msgHash, err := senderNode.Wrappers_RelayPublish(defaultPubsubTopic, message)
	require.NoError(t, err)
	require.NotEmpty(t, msgHash)

	logger.Debug("Waiting to ensure message delivery")
	time.Sleep(2 * time.Second)

	logger.Debug("Verifying message reception using the new wrapper")
	err = receiverNode.Wrappers_VerifyMessageReceived(message, msgHash)
	require.NoError(t, err, "message verification failed")

	logger.Debug("TestRelayMessageTransmission completed successfully")
}
