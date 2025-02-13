package waku

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/waku-org/go-waku/waku/v2/protocol/pb"
	"github.com/waku-org/waku-go-bindings/waku/common"
	"google.golang.org/protobuf/proto"
)

func TestRelaySubscribeToDefaultTopic(t *testing.T) {
	Debug("Starting test to verify relay subscription to the default pubsub topic")

	wakuConfig := DefaultWakuConfig
	wakuConfig.Relay = true

	Debug("Creating a Waku node with relay enabled")
	node, err := StartWakuNode("TestNode", &wakuConfig)
	require.NoError(t, err)
	defer func() {
		Debug("Stopping and destroying the Waku node")
		node.StopAndDestroy()
	}()

	defaultPubsubTopic := DefaultPubsubTopic
	Debug("Default pubsub topic retrieved: %s", defaultPubsubTopic)

	Debug("Attempting to subscribe to the default pubsub topic %s", defaultPubsubTopic)
	err = RetryWithBackOff(func() error {
		return node.RelaySubscribe(defaultPubsubTopic)
	})
	require.NoError(t, err)

	Debug("Test successfully verified subscription to the default pubsub topic: %s", defaultPubsubTopic)
}

func TestRelayMessageTransmission(t *testing.T) {
	Debug("Starting TestRelayMessageTransmission")

	senderConfig := DefaultWakuConfig
	senderConfig.Relay = true
	senderNode, err := StartWakuNode("SenderNode", &senderConfig)
	require.NoError(t, err)
	defer senderNode.StopAndDestroy()

	receiverConfig := DefaultWakuConfig
	receiverConfig.Relay = true
	receiverNode, err := StartWakuNode("ReceiverNode", &receiverConfig)
	require.NoError(t, err)
	defer receiverNode.StopAndDestroy()

	Debug("Connecting sender and receiver")
	err = RetryWithBackOff(func() error {
		return senderNode.ConnectPeer(receiverNode)
	})
	require.NoError(t, err)

	Debug("Subscribing receiver to the default pubsub topic")
	defaultPubsubTopic := DefaultPubsubTopic
	err = RetryWithBackOff(func() error {
		return receiverNode.RelaySubscribe(defaultPubsubTopic)
	})
	require.NoError(t, err)

	Debug("Creating and publishing message")
	message := senderNode.CreateMessage()
	var msgHash string
	err = RetryWithBackOff(func() error {
		var err error
		msgHashObj, err := senderNode.RelayPublishNoCTX(defaultPubsubTopic, message)
		if err == nil {
			msgHash = msgHashObj.String()
		}
		return err

	})
	require.NoError(t, err)
	require.NotEmpty(t, msgHash)

	Debug("Waiting to ensure message delivery")
	time.Sleep(2 * time.Second)

	Debug("Verifying message reception")
	err = RetryWithBackOff(func() error {
		msgHashObj, _ := common.ToMessageHash(msgHash)
		return receiverNode.VerifyMessageReceived(message, msgHashObj)
	})
	require.NoError(t, err, "message verification failed")

	Debug("TestRelayMessageTransmission completed successfully")
}

func TestRelayMessageBroadcast(t *testing.T) {
	Debug("Starting TestRelayMessageBroadcast")

	numPeers := 5
	nodes := make([]*WakuNode, numPeers)
	nodeNames := []string{"SenderNode", "PeerNode1", "PeerNode2", "PeerNode3", "PeerNode4"}
	defaultPubsubTopic := DefaultPubsubTopic

	for i := 0; i < numPeers; i++ {
		Debug("Creating node %s", nodeNames[i])
		nodeConfig := DefaultWakuConfig
		nodeConfig.Relay = true
		node, err := StartWakuNode(nodeNames[i], &nodeConfig)
		require.NoError(t, err)
		defer node.StopAndDestroy()
		nodes[i] = node
	}

	err := ConnectAllPeers(nodes)
	require.NoError(t, err)

	Debug("Subscribing nodes to the default pubsub topic")
	for _, node := range nodes {
		err := node.RelaySubscribe(defaultPubsubTopic)
		require.NoError(t, err)
	}

	senderNode := nodes[0]
	Debug("SenderNode is publishing a message")
	message := senderNode.CreateMessage()
	msgHash, err := senderNode.RelayPublishNoCTX(defaultPubsubTopic, message)
	require.NoError(t, err)
	require.NotEmpty(t, msgHash)

	Debug("Waiting to ensure message delivery")
	time.Sleep(3 * time.Second)

	Debug("Verifying message reception for each node")
	for i, node := range nodes {
		Debug("Verifying message for node %s", nodeNames[i])
		err := node.VerifyMessageReceived(message, msgHash)
		require.NoError(t, err, "message verification failed for node: %s", nodeNames[i])
	}

	Debug("TestRelayMessageBroadcast completed successfully")
}

func TestSendmsgInvalidPayload(t *testing.T) {
	Debug("Starting TestInvalidMessageFormat")

	nodeNames := []string{"SenderNode", "PeerNode1"}
	defaultPubsubTopic := DefaultPubsubTopic

	Debug("Creating nodes")
	senderNodeConfig := DefaultWakuConfig
	senderNodeConfig.Relay = true
	senderNode, err := StartWakuNode(nodeNames[0], &senderNodeConfig)
	require.NoError(t, err)
	defer senderNode.StopAndDestroy()

	receiverNodeConfig := DefaultWakuConfig
	receiverNodeConfig.Relay = true
	receiverNode, err := StartWakuNode(nodeNames[1], &receiverNodeConfig)
	require.NoError(t, err)
	defer receiverNode.StopAndDestroy()

	Debug("Connecting SenderNode and PeerNode1")
	err = senderNode.ConnectPeer(receiverNode)
	require.NoError(t, err)

	Debug("Subscribing SenderNode to the default pubsub topic")
	err = senderNode.RelaySubscribe(defaultPubsubTopic)
	require.NoError(t, err)

	Debug("SenderNode is publishing an invalid message")
	invalidMessage := &pb.WakuMessage{
		Payload: []byte{}, // Empty payload
		Version: proto.Uint32(0),
	}

	message := senderNode.CreateMessage(invalidMessage)
	msgHash, err := senderNode.RelayPublishNoCTX(defaultPubsubTopic, message)

	Debug("Verifying if message was sent or failed")
	if err != nil {
		Debug("Message was not sent due to invalid format: %v", err)
		require.Error(t, err, "message should fail due to invalid format")
	} else {
		Debug("Message was unexpectedly sent: %s", msgHash.String())
		require.Fail(t, "message with invalid format should not be sent")
	}

	Debug("TestInvalidMessageFormat completed")
}

func TestMessageNotReceivedWithoutRelay(t *testing.T) {
	Debug("Starting TestMessageNotReceivedWithoutRelay")

	Debug("Creating Sender Node with Relay disabled")
	senderConfig := DefaultWakuConfig
	senderConfig.Relay = true
	senderNode, err := StartWakuNode("SenderNode", &senderConfig)
	require.NoError(t, err)
	defer senderNode.StopAndDestroy()

	Debug("Creating Receiver Node with Relay disabled")
	receiverConfig := DefaultWakuConfig
	receiverConfig.Relay = true
	receiverNode, err := StartWakuNode("ReceiverNode", &receiverConfig)
	require.NoError(t, err)
	defer receiverNode.StopAndDestroy()

	Debug("Connecting Sender and Receiver")
	err = senderNode.ConnectPeer(receiverNode)
	require.NoError(t, err)

	Debug("Subscribing Receiver to the default pubsub topic")
	defaultPubsubTopic := DefaultPubsubTopic
	err = senderNode.RelaySubscribe(defaultPubsubTopic)
	require.NoError(t, err)
	err = receiverNode.RelaySubscribe(defaultPubsubTopic)
	require.NoError(t, err)

	Debug("SenderNode is publishing a message")
	message := senderNode.CreateMessage()
	msgHash, err := senderNode.RelayPublishNoCTX(defaultPubsubTopic, message)
	require.NoError(t, err)
	require.NotEmpty(t, msgHash)

	Debug("Waiting to ensure message delivery")
	time.Sleep(3 * time.Second)

	Debug("Verifying that Receiver did NOT receive the message")
	err = receiverNode.VerifyMessageReceived(message, msgHash)
	if err == nil {
		t.Fatalf("Test failed: ReceiverNode SHOULD NOT have received the message")
	}

	Debug("TestMessageNotReceivedWithoutRelay completed successfully")
}
