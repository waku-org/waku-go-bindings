package waku_go_tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/waku-org/go-waku/waku/v2/protocol/pb"
	testlibs "github.com/waku-org/waku-go-bindings/testlibs/src"
	utilities "github.com/waku-org/waku-go-bindings/testlibs/utilities"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
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

func TestRelayMessageBroadcast(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	logger.Debug("Starting TestRelayMessageBroadcast")

	numPeers := 5
	nodes := make([]*testlibs.WakuNodeWrapper, numPeers)
	nodeNames := []string{"SenderNode", "PeerNode1", "PeerNode2", "PeerNode3", "PeerNode4"}

	defaultPubsubTopic := utilities.DefaultPubsubTopic

	for i := 0; i < numPeers; i++ {
		logger.Debug("Creating node", zap.String("node", nodeNames[i]))

		nodeConfig := *utilities.DefaultWakuConfig
		nodeConfig.Relay = true

		node, err := testlibs.Wrappers_StartWakuNode(&nodeConfig, logger.Named(nodeNames[i]))
		require.NoError(t, err)
		defer node.Wrappers_StopAndDestroy()

		nodes[i] = node
	}

	err = testlibs.Wrappers_ConnectAllPeers(nodes)
	require.NoError(t, err)

	logger.Debug("Subscribing nodes to the default pubsub topic")
	for _, node := range nodes {
		err := node.Wrappers_RelaySubscribe(defaultPubsubTopic)
		require.NoError(t, err)
	}

	senderNode := nodes[0]
	logger.Debug("SenderNode is publishing a message")
	message := senderNode.Wrappers_CreateMessage()
	msgHash, err := senderNode.Wrappers_RelayPublish(defaultPubsubTopic, message)
	require.NoError(t, err)
	require.NotEmpty(t, msgHash)

	logger.Debug("Waiting to ensure message delivery")
	time.Sleep(3 * time.Second)

	logger.Debug("Verifying message reception for each node")
	for i, node := range nodes {
		logger.Debug("Verifying message for node", zap.String("node", nodeNames[i]))
		err := node.Wrappers_VerifyMessageReceived(message, msgHash)
		require.NoError(t, err, "message verification failed for node: "+nodeNames[i])
	}

	logger.Debug("TestRelayMessageBroadcast completed successfully")
}

func TestSendmsgInvalidPayload(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	logger.Debug("Starting TestInvalidMessageFormat")

	nodeNames := []string{"SenderNode", "PeerNode1"}

	defaultPubsubTopic := utilities.DefaultPubsubTopic

	logger.Debug("Creating nodes")
	senderNodeConfig := *utilities.DefaultWakuConfig
	senderNodeConfig.Relay = true
	senderNode, err := testlibs.Wrappers_StartWakuNode(&senderNodeConfig, logger.Named(nodeNames[0]))
	require.NoError(t, err)
	defer senderNode.Wrappers_StopAndDestroy()

	receiverNodeConfig := *utilities.DefaultWakuConfig
	receiverNodeConfig.Relay = true

	receiverNode, err := testlibs.Wrappers_StartWakuNode(&receiverNodeConfig, logger.Named(nodeNames[1]))
	require.NoError(t, err)
	defer receiverNode.Wrappers_StopAndDestroy()

	logger.Debug("Connecting SenderNode and PeerNode1")
	err = senderNode.Wrappers_ConnectPeer(receiverNode)
	require.NoError(t, err)

	logger.Debug("Subscribing SenderNode to the default pubsub topic")
	err = senderNode.Wrappers_RelaySubscribe(defaultPubsubTopic)
	require.NoError(t, err)

	logger.Debug("SenderNode is publishing an invalid message")
	invalidMessage := &pb.WakuMessage{
		Payload: []byte{}, // Empty payload
		Version: proto.Uint32(0),
	}

	message := senderNode.Wrappers_CreateMessage(invalidMessage)

	msgHash, err := senderNode.Wrappers_RelayPublish(defaultPubsubTopic, message)

	logger.Debug("Verifying if message was sent or failed")
	if err != nil {
		logger.Debug("Message was not sent due to invalid format", zap.Error(err))
		require.Error(t, err, "message should fail due to invalid format")
	} else {
		logger.Debug("Message was unexpectedly sent", zap.String("messageHash", msgHash.String()))
		require.Fail(t, "message with invalid format should not be sent")
	}

	logger.Debug("TestInvalidMessageFormat completed")
}
