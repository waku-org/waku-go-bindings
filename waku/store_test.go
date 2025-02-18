package waku

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/waku-org/go-waku/waku/v2/protocol/pb"

	"github.com/waku-org/waku-go-bindings/waku/common"
	"google.golang.org/protobuf/proto"
)

func TestStoreQueryFromPeer(t *testing.T) {
	Debug("Starting test to verify store query from a peer using direct peer connections")

	node1Config := DefaultWakuConfig
	node1Config.Relay = true

	Debug("Creating Node1 (Relay enabled)")
	node1, err := StartWakuNode("Node1", &node1Config)
	require.NoError(t, err, "Failed to start Node1")

	node2Config := DefaultWakuConfig
	node2Config.Relay = true
	node2Config.Store = true

	Debug("Creating Node2 (Relay & Store enabled)")
	node2, err := StartWakuNode("Node2", &node2Config)
	require.NoError(t, err, "Failed to start Node2")

	node3Config := DefaultWakuConfig
	node3Config.Relay = false

	Debug("Creating Node3 (Peer connected to Node2)")
	node3, err := StartWakuNode("Node3", &node3Config)
	require.NoError(t, err, "Failed to start Node3")

	defer func() {
		Debug("Stopping and destroying all Waku nodes")
		node1.StopAndDestroy()
		node2.StopAndDestroy()
		node3.StopAndDestroy()
	}()

	Debug("Connecting Node1 to Node2")
	err = node1.ConnectPeer(node2)
	require.NoError(t, err, "Failed to connect Node1 to Node2")

	Debug("Connecting Node3 to Node2")
	err = node3.ConnectPeer(node2)
	require.NoError(t, err, "Failed to connect Node3 to Node2")

	Debug("Waiting for peer connections to stabilize")
	time.Sleep(2 * time.Second)

	Debug("Publishing message from Node1 using RelayPublish")
	message := node1.CreateMessage(&pb.WakuMessage{
		Payload:      []byte("test-message"),
		ContentTopic: "test-content-topic",
		Timestamp:    proto.Int64(time.Now().UnixNano()),
	})

	defaultPubsubTopic := DefaultPubsubTopic
	msgHash, err := node1.RelayPublishNoCTX(defaultPubsubTopic, message)
	require.NoError(t, err, "Failed to publish message from Node1")

	Debug("Waiting for message delivery to Node2")
	time.Sleep(2 * time.Second)

	Debug("Verifying that Node2 received the message")
	err = node2.VerifyMessageReceived(message, msgHash)
	require.NoError(t, err, "Node2 should have received the message")

	Debug("Node3 querying stored messages from Node2")
	res, err := node3.GetStoredMessages(node2, nil)
	var storedMessages = *res.Messages
	require.NoError(t, err, "Failed to retrieve stored messages from Node2")
	require.NotEmpty(t, storedMessages, "Expected at least one stored message")
	Debug("Verifying stored message matches the published message")
	require.Equal(t, message.Payload, storedMessages[0].WakuMessage.Payload, "Stored message payload does not match")
	Debug("Test successfully verified store query from a peer using direct peer connections")
}

func TestStoreQueryMultipleMessages(t *testing.T) {
	Debug("Starting test to verify store query with multiple messages")

	node1Config := DefaultWakuConfig
	node1Config.Relay = true

	Debug("Creating Node1 (Relay enabled)")
	node1, err := StartWakuNode("Node1", &node1Config)
	require.NoError(t, err, "Failed to start Node1")

	node2Config := DefaultWakuConfig
	node2Config.Relay = true
	node2Config.Store = true

	Debug("Creating Node2 (Relay & Store enabled)")
	node2, err := StartWakuNode("Node2", &node2Config)
	require.NoError(t, err, "Failed to start Node2")

	node3Config := DefaultWakuConfig
	node3Config.Relay = false

	Debug("Creating Node3 (Peer connected to Node2)")
	node3, err := StartWakuNode("Node3", &node3Config)
	require.NoError(t, err, "Failed to start Node3")

	defer func() {
		Debug("Stopping and destroying all Waku nodes")
		node1.StopAndDestroy()
		node2.StopAndDestroy()
		node3.StopAndDestroy()
	}()

	Debug("Connecting Node1 to Node2")
	err = node1.ConnectPeer(node2)
	require.NoError(t, err, "Failed to connect Node1 to Node2")

	Debug("Connecting Node3 to Node2")
	err = node3.ConnectPeer(node2)
	require.NoError(t, err, "Failed to connect Node3 to Node2")

	Debug("Waiting for peer connections to stabilize")
	time.Sleep(2 * time.Second)

	numMessages := 50
	var sentHashes []common.MessageHash
	defaultPubsubTopic := DefaultPubsubTopic

	Debug("Publishing %d messages from Node1 using RelayPublish", numMessages)
	for i := 0; i < numMessages; i++ {
		message := node1.CreateMessage(&pb.WakuMessage{
			Payload:      []byte(fmt.Sprintf("message-%d", i)),
			ContentTopic: "test-content-topic",
			Timestamp:    proto.Int64(time.Now().UnixNano()),
		})

		msgHash, err := node1.RelayPublishNoCTX(defaultPubsubTopic, message)
		require.NoError(t, err, "Failed to publish message from Node1")

		sentHashes = append(sentHashes, msgHash)
	}

	Debug("Waiting for message delivery to Node2")
	time.Sleep(5 * time.Second)

	Debug("Node3 querying stored messages from Node2")
	res, err := node3.GetStoredMessages(node2, nil)
	require.NoError(t, err, "Failed to retrieve stored messages from Node2")
	require.NotNil(t, res.Messages, "Expected stored messages but received nil")

	storedMessages := *res.Messages

	require.Len(t, storedMessages, numMessages, "Expected to retrieve exactly 50 messages")

	Debug("Verifying stored message hashes match sent message hashes")
	var receivedHashes []common.MessageHash
	for _, storedMsg := range storedMessages {
		receivedHashes = append(receivedHashes, storedMsg.MessageHash)
	}

	require.ElementsMatch(t, sentHashes, receivedHashes, "Sent and received message hashes do not match")

	Debug("Test successfully verified store query with multiple messages")
}

func TestStoreQueryWith5Pagination(t *testing.T) {
	Debug("Starting test to verify store query with pagination")

	node1Config := DefaultWakuConfig
	node1Config.Relay = true

	Debug("Creating Node1 (Relay enabled)")
	node1, err := StartWakuNode("Node1", &node1Config)
	require.NoError(t, err, "Failed to start Node1")

	node2Config := DefaultWakuConfig
	node2Config.Relay = true
	node2Config.Store = true

	Debug("Creating Node2 (Relay & Store enabled)")
	node2, err := StartWakuNode("Node2", &node2Config)
	require.NoError(t, err, "Failed to start Node2")

	node3Config := DefaultWakuConfig
	node3Config.Relay = false

	Debug("Creating Node3 (Peer connected to Node2)")
	node3, err := StartWakuNode("Node3", &node3Config)
	require.NoError(t, err, "Failed to start Node3")

	defer func() {
		Debug("Stopping and destroying all Waku nodes")
		node1.StopAndDestroy()
		node2.StopAndDestroy()
		node3.StopAndDestroy()
	}()

	Debug("Connecting Node1 to Node2")
	err = node1.ConnectPeer(node2)
	require.NoError(t, err, "Failed to connect Node1 to Node2")

	Debug("Connecting Node3 to Node2")
	err = node3.ConnectPeer(node2)
	require.NoError(t, err, "Failed to connect Node3 to Node2")

	Debug("Waiting for peer connections to stabilize")
	time.Sleep(2 * time.Second)

	numMessages := 10
	defaultPubsubTopic := DefaultPubsubTopic

	Debug("Publishing %d messages from Node1 using RelayPublish", numMessages)
	for i := 0; i < numMessages; i++ {
		message := node1.CreateMessage(&pb.WakuMessage{
			Payload:      []byte(fmt.Sprintf("message-%d", i)),
			ContentTopic: "test-content-topic",
			Timestamp:    proto.Int64(time.Now().UnixNano()),
		})

		_, err := node1.RelayPublishNoCTX(defaultPubsubTopic, message)
		require.NoError(t, err, "Failed to publish message from Node1")

	}

	Debug("Waiting for message delivery to Node2")
	time.Sleep(5 * time.Second)

	Debug("Node3 querying stored messages from Node2 with PaginationLimit = 5")
	storeRequest := common.StoreQueryRequest{
		IncludeData:       true,
		ContentTopics:     &[]string{"test-content-topic"},
		PaginationLimit:   proto.Uint64(uint64(5)),
		PaginationForward: true,
	}

	res, err := node3.GetStoredMessages(node2, &storeRequest)
	require.NoError(t, err, "Failed to retrieve stored messages from Node2")
	require.NotNil(t, res.Messages, "Expected stored messages but received nil")

	storedMessages := *res.Messages

	require.Len(t, storedMessages, 5, "Expected to retrieve exactly 5 messages due to pagination limit")

	Debug("Test successfully verified store query with pagination limit")
}
