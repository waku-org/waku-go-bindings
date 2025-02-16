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
	Debug("Starting test to verify relay subscription with multiple nodes via Discv5")

	// Configure Node1
	node1Config := DefaultWakuConfig
	node1Config.Relay = true

	Debug("Creating Node1 with Relay enabled")
	node1, err := StartWakuNode("Node1", &node1Config)
	require.NoError(t, err, "Failed to start Node1")

	// Retrieve Node1's ENR for Discv5 bootstrapping
	enrNode1, err := node1.ENR()
	require.NoError(t, err, "Failed to get ENR for Node1")

	// Configure Node2 with Node1 as its Discv5 bootstrap node
	node2Config := DefaultWakuConfig
	node2Config.Relay = true
	node2Config.Discv5BootstrapNodes = []string{enrNode1.String()}

	Debug("Creating Node2 with Node1 as Discv5 bootstrap")
	node2, err := StartWakuNode("Node2", &node2Config)
	require.NoError(t, err, "Failed to start Node2")

	// Configure Node3 with Node1 as its Discv5 bootstrap node
	node3Config := DefaultWakuConfig
	node3Config.Relay = true
	node3Config.Discv5BootstrapNodes = []string{enrNode1.String()}

	Debug("Creating Node3 with Node1 as Discv5 bootstrap")
	node3, err := StartWakuNode("Node3", &node3Config)
	require.NoError(t, err, "Failed to start Node3")

	defer func() {
		Debug("Stopping and destroying all Waku nodes")
		node1.StopAndDestroy()
		node2.StopAndDestroy()
		node3.StopAndDestroy()
	}()

	defaultPubsubTopic := DefaultPubsubTopic
	Debug("Default pubsub topic retrieved: %s", defaultPubsubTopic)

	Debug("Waiting for nodes to auto-connect via Discv5")
	err = WaitForAutoConnection([]*WakuNode{node1, node2, node3})
	require.NoError(t, err, "Nodes did not auto-connect within timeout")

	Debug("Subscribing Node1 to the default pubsub topic: %s", defaultPubsubTopic)
	err = RetryWithBackOff(func() error {
		return node1.RelaySubscribe(defaultPubsubTopic)
	})
	require.NoError(t, err)

	Debug("Subscribing Node2 to the default pubsub topic: %s", defaultPubsubTopic)
	err = RetryWithBackOff(func() error {
		return node2.RelaySubscribe(defaultPubsubTopic)
	})
	require.NoError(t, err)

	Debug("Subscribing Node3 to the default pubsub topic: %s", defaultPubsubTopic)
	err = RetryWithBackOff(func() error {
		return node3.RelaySubscribe(defaultPubsubTopic)
	})
	require.NoError(t, err)

	Debug("Waiting for peer connections to stabilize")
	time.Sleep(10 * time.Second)

	Debug("Fetching number of connected peers for Node1")
	peers, err := node1.GetNumConnectedRelayPeers(defaultPubsubTopic)
	require.NoError(t, err, "Failed to get number of connected relay peers for Node1")

	Debug("Total number of connected relay peers for Node1: %d", peers)

	if peers != 2 {
		Error("Expected Node1 to have exactly 2 relay peers, but got %d", peers)
		t.FailNow()
	}

	Debug("Test successfully verified that Node1 has 2 relay peers after Node2 and Node3 subscribed")
}

func TestRelayMessageTransmission(t *testing.T) {
	Debug("Starting TestRelayMessageTransmission")

	Debug("Creating Sender Node with Relay enabled")
	senderConfig := DefaultWakuConfig
	senderConfig.Relay = true

	senderNode, err := StartWakuNode("SenderNode", &senderConfig)
	require.NoError(t, err, "Failed to start SenderNode")
	defer senderNode.StopAndDestroy()

	Debug("Creating Receiver Node with Relay enabled")
	receiverConfig := DefaultWakuConfig
	receiverConfig.Relay = true

	// Set the Receiver Node's discovery bootstrap node as SenderNode
	enrSender, err := senderNode.ENR()
	require.NoError(t, err, "Failed to get ENR for SenderNode")
	receiverConfig.Discv5BootstrapNodes = []string{enrSender.String()}

	receiverNode, err := StartWakuNode("ReceiverNode", &receiverConfig)
	require.NoError(t, err, "Failed to start ReceiverNode")
	defer receiverNode.StopAndDestroy()

	Debug("Waiting for nodes to auto-connect via Discv5")
	err = WaitForAutoConnection([]*WakuNode{senderNode, receiverNode})
	require.NoError(t, err, "Nodes did not auto-connect within timeout")

	Debug("Subscribing ReceiverNode to the default pubsub topic")
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
	require.NoError(t, err, "Message verification failed")

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
		if i > 0 {
			enrPrevNode, err := nodes[i-1].ENR()
			require.NoError(t, err, "Failed to get ENR for node %s", nodeNames[i-1])
			nodeConfig.Discv5BootstrapNodes = []string{enrPrevNode.String()}
		}
		node, err := StartWakuNode(nodeNames[i], &nodeConfig)
		require.NoError(t, err)
		defer node.StopAndDestroy()
		nodes[i] = node
	}

	Debug("Subscribing nodes to the default pubsub topic")
	for _, node := range nodes {
		err := node.RelaySubscribe(defaultPubsubTopic)
		require.NoError(t, err)
	}

	WaitForAutoConnection(nodes)

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

	defaultPubsubTopic := DefaultPubsubTopic

	Debug("Creating nodes")
	senderNodeConfig := DefaultWakuConfig
	senderNodeConfig.Relay = true
	senderNode, err := StartWakuNode("SenderNode", &senderNodeConfig)
	require.NoError(t, err)
	defer senderNode.StopAndDestroy()

	receiverNodeConfig := DefaultWakuConfig
	receiverNodeConfig.Relay = true
	enrNode2, err := senderNode.ENR()
	if err != nil {
		require.Error(t, err, "Can't find node ENR")
	}
	receiverNodeConfig.Discv5BootstrapNodes = []string{enrNode2.String()}
	receiverNode, err := StartWakuNode("receiverNode", &receiverNodeConfig)
	require.NoError(t, err)
	defer receiverNode.StopAndDestroy()

	Debug("Subscribing SenderNode to the default pubsub topic")
	err = senderNode.RelaySubscribe(defaultPubsubTopic)
	require.NoError(t, err)

	err = WaitForAutoConnection([]*WakuNode{senderNode, receiverNode})
	require.NoError(t, err, "Nodes did not auto-connect within timeout")

	Debug("SenderNode is publishing an invalid message")
	invalidMessage := &pb.WakuMessage{
		Payload:      []byte{},
		ContentTopic: "test-content-topic",
		Version:      proto.Uint32(0),
		Timestamp:    proto.Int64(time.Now().UnixNano()),
	}

	message := senderNode.CreateMessage(invalidMessage)
	var msgHash common.MessageHash
	msgHash, err = senderNode.RelayPublishNoCTX(defaultPubsubTopic, message)

	Debug("Verifying if message was sent or failed")
	if err != nil {
		Debug("Message was not sent due to invalid format: %v", err)

	} else {
		Debug("Message was unexpectedly sent: %s", msgHash.String())
		require.Fail(t, "message with invalid format should not be sent")
	}

	Debug("TestInvalidMessageFormat completed")
}

func TestRelayNodesNotConnectedDirectly(t *testing.T) {
	Debug("Starting TestMessageNotReceivedWithoutRelay")

	Debug("Creating Sender Node with Relay enabled")
	senderConfig := DefaultWakuConfig
	senderConfig.Relay = true
	senderNode, err := StartWakuNode("SenderNode", &senderConfig)
	require.NoError(t, err)
	defer senderNode.StopAndDestroy()

	Debug("Creating Relay-Enabled Receiver Node (Node2)")
	node2Config := DefaultWakuConfig
	node2Config.Relay = true
	enrNode, err := senderNode.ENR()
	if err != nil {
		require.Error(t, err, "Can't find node ENR")
	}
	node2Config.Discv5BootstrapNodes = []string{enrNode.String()}
	node2, err := StartWakuNode("Node2", &node2Config)
	require.NoError(t, err)
	defer node2.StopAndDestroy()

	Debug("Creating Non-Relay Receiver Node (Node3)")
	node3Config := DefaultWakuConfig
	node3Config.Relay = true
	enrNode2, err := node2.ENR()
	if err != nil {
		require.Error(t, err, "Can't find node ENR")
	}
	node3Config.Discv5BootstrapNodes = []string{enrNode2.String()}
	node3, err := StartWakuNode("Node3", &node3Config)
	require.NoError(t, err)
	defer node3.StopAndDestroy()

	Debug("Subscribing Node2 to the default pubsub topic")
	defaultPubsubTopic := DefaultPubsubTopic
	err = node2.RelaySubscribe(defaultPubsubTopic)
	require.NoError(t, err)

	Debug("Subscribing Node2 to the default pubsub topic")
	err = node3.RelaySubscribe(defaultPubsubTopic)
	require.NoError(t, err)

	Debug("Waiting for nodes to auto-connect before proceeding")
	err = WaitForAutoConnection([]*WakuNode{senderNode, node2, node3})
	require.NoError(t, err, "Nodes did not auto-connect within timeout")

	Debug("SenderNode is publishing a message")
	message := senderNode.CreateMessage()
	msgHash, err := senderNode.RelayPublishNoCTX(defaultPubsubTopic, message)
	require.NoError(t, err)
	require.NotEmpty(t, msgHash)

	Debug("Verifying that Node2 received the message")
	err = node2.VerifyMessageReceived(message, msgHash)
	require.NoError(t, err, "Node2 should have received the message")

	Debug("Verifying that Node3 receive the message")
	err = node3.VerifyMessageReceived(message, msgHash)
	require.NoError(t, err, "Node3 should  have received the message")

	Debug("TestMessageNotReceivedWithoutRelay completed successfully")
}

func TestRelaySubscribeAndPeerCountChange(t *testing.T) {
	Debug("Starting test to verify relay subscription and peer count change after stopping a node")

	node1Config := DefaultWakuConfig
	node1Config.Relay = true

	Debug("Creating Node1 with Relay enabled")
	node1, err := StartWakuNode("Node1", &node1Config)
	require.NoError(t, err, "Failed to start Node1")

	enrNode1, err := node1.ENR()
	require.NoError(t, err, "Failed to get ENR for Node1")

	node2Config := DefaultWakuConfig
	node2Config.Relay = true
	node2Config.Discv5BootstrapNodes = []string{enrNode1.String()}

	Debug("Creating Node2 with Node1 as Discv5 bootstrap")
	node2, err := StartWakuNode("Node2", &node2Config)
	require.NoError(t, err, "Failed to start Node2")

	//enrNode2, err := node2.ENR()
	//require.NoError(t, err, "Failed to get ENR for Node2")

	node3Config := DefaultWakuConfig
	node3Config.Relay = true
	node3Config.Discv5BootstrapNodes = []string{enrNode1.String()}

	Debug("Creating Node3 with Node2 as Discv5 bootstrap")
	node3, err := StartWakuNode("Node3", &node3Config)
	require.NoError(t, err, "Failed to start Node3")

	defer func() {
		Debug("Stopping and destroying all Waku nodes")
		node1.StopAndDestroy()
		node2.StopAndDestroy()
	}()

	defaultPubsubTopic := DefaultPubsubTopic
	Debug("Default pubsub topic retrieved: %s", defaultPubsubTopic)

	Debug("Waiting for nodes to auto-connect via Discv5")
	err = WaitForAutoConnection([]*WakuNode{node1, node2, node3})
	require.NoError(t, err, "Nodes did not auto-connect within timeout")

	Debug("Subscribing Node1 to the default pubsub topic: %s", defaultPubsubTopic)
	err = RetryWithBackOff(func() error {
		return node1.RelaySubscribe(defaultPubsubTopic)
	})
	require.NoError(t, err)

	Debug("Subscribing Node2 to the default pubsub topic: %s", defaultPubsubTopic)
	err = RetryWithBackOff(func() error {
		return node2.RelaySubscribe(defaultPubsubTopic)
	})
	require.NoError(t, err)

	Debug("Subscribing Node3 to the default pubsub topic: %s", defaultPubsubTopic)
	err = RetryWithBackOff(func() error {
		return node3.RelaySubscribe(defaultPubsubTopic)
	})
	require.NoError(t, err)

	Debug("Waiting for peer connections to stabilize")
	time.Sleep(10 * time.Second)

	Debug("Fetching number of connected peers for Node1")
	peersBeforeStopping, err := node1.GetNumConnectedRelayPeers(defaultPubsubTopic)
	require.NoError(t, err, "Failed to get number of connected relay peers for Node1")

	Debug("Total number of connected relay peers for Node1 before stopping Node3: %d", peersBeforeStopping)
	require.Equal(t, 2, peersBeforeStopping, "Expected Node1 to have exactly 2 relay peers before stopping Node3")
	Debug("Stopping Node3")
	node3.StopAndDestroy()

	Debug("Waiting for network to update after Node3 stops")
	time.Sleep(10 * time.Second)

	Debug("Fetching number of connected peers for Node1 after stopping Node3")
	peersAfterStopping, err := node1.GetNumConnectedRelayPeers(defaultPubsubTopic)
	require.NoError(t, err, "Failed to get number of connected relay peers for Node1 after stopping Node3")

	Debug("Total number of connected relay peers for Node1 after stopping Node3: %d", peersAfterStopping)

	require.Equal(t, 1, peersAfterStopping, "Expected Node1 to have exactly 1 relay peer after stopping Node3")
	Debug("Test successfully verified peer count changes as expected after stopping Node3")
}

func TestDiscv5PeerMeshCount(t *testing.T) {
	Debug("Starting test to verify peer count in mesh using Discv5 after topic subscription")

	node1Config := DefaultWakuConfig
	node1Config.Relay = true
	Debug("Creating Node1")
	node1, err := StartWakuNode("Node1", &node1Config)
	require.NoError(t, err, "Failed to start Node1")

	enrNode1, err := node1.ENR()
	require.NoError(t, err, "Failed to get ENR for Node1")

	node2Config := DefaultWakuConfig
	node2Config.Discv5BootstrapNodes = []string{enrNode1.String()}
	node2Config.Relay = true
	Debug("Creating Node2 with Node1 as Discv5 bootstrap")
	node2, err := StartWakuNode("Node2", &node2Config)
	require.NoError(t, err, "Failed to start Node2")

	require.NoError(t, err, "Failed to get ENR for Node2")

	node3Config := DefaultWakuConfig
	node3Config.Discv5BootstrapNodes = []string{enrNode1.String()}
	node3Config.Relay = true

	Debug("Creating Node3 with Node2 as Discv5 bootstrap")
	node3, err := StartWakuNode("Node3", &node3Config)
	require.NoError(t, err, "Failed to start Node3")

	defer func() {
		Debug("Stopping and destroying all Waku nodes")
		node1.StopAndDestroy()
		node2.StopAndDestroy()
	}()

	defaultPubsubTopic := DefaultPubsubTopic
	Debug("Default pubsub topic retrieved: %s", defaultPubsubTopic)

	err = SubscribeNodesToTopic([]*WakuNode{node1, node2, node3}, defaultPubsubTopic)
	require.NoError(t, err, "Failed to subscribe all nodes to the topic")

	Debug("Waiting for nodes to auto-connect via Discv5")
	err = WaitForAutoConnection([]*WakuNode{node1, node2, node3})
	require.NoError(t, err, "Nodes did not auto-connect within timeout")

	Debug("Fetching number of peers in mesh for Node1 before stopping Node3")
	peerCountBefore, err := node1.GetNumPeersInMesh(defaultPubsubTopic)
	require.NoError(t, err, "Failed to get number of peers in mesh for Node1 before stopping Node3")

	Debug("Total number of peers in mesh for Node1 before stopping Node3: %d", peerCountBefore)
	require.Equal(t, 2, peerCountBefore, "Expected Node1 to have exactly 3 peers in the mesh before stopping Node3")

	Debug("Stopping Node3")
	node3.StopAndDestroy()

	Debug("Waiting for network update after Node3 stops")
	time.Sleep(10 * time.Second)

	Debug("Fetching number of peers in mesh for Node1 after stopping Node3")
	peerCountAfter, err := node1.GetNumPeersInMesh(defaultPubsubTopic)
	require.NoError(t, err, "Failed to get number of peers in mesh for Node1 after stopping Node3")

	Debug("Total number of peers in mesh for Node1 after stopping Node3: %d", peerCountAfter)
	require.Equal(t, 1, peerCountAfter, "Expected Node1 to have exactly 2 peers in the mesh after stopping Node3")

	Debug("Test successfully verified peer count change after stopping Node3")
}

func TestDiscv5GetPeersConnected(t *testing.T) {
	Debug("Starting test to verify peer count in mesh with 4 nodes using Discv5 (Chained Connection)")

	node1Config := DefaultWakuConfig
	node1Config.Relay = true

	Debug("Creating Node1")
	node1, err := StartWakuNode("Node1", &node1Config)
	require.NoError(t, err, "Failed to start Node1")

	enrNode1, err := node1.ENR()
	require.NoError(t, err, "Failed to get ENR for Node1")

	node2Config := DefaultWakuConfig
	node2Config.Relay = true
	node2Config.Discv5BootstrapNodes = []string{enrNode1.String()}

	Debug("Creating Node2 with Node1 as Discv5 bootstrap")
	node2, err := StartWakuNode("Node2", &node2Config)
	require.NoError(t, err, "Failed to start Node2")

	enrNode2, err := node2.ENR()
	require.NoError(t, err, "Failed to get ENR for Node2")

	node3Config := DefaultWakuConfig
	node3Config.Relay = true
	node3Config.Discv5BootstrapNodes = []string{enrNode2.String()}

	Debug("Creating Node3 with Node2 as Discv5 bootstrap")
	node3, err := StartWakuNode("Node3", &node3Config)
	require.NoError(t, err, "Failed to start Node3")

	enrNode3, err := node3.ENR()
	require.NoError(t, err, "Failed to get ENR for Node3")

	node4Config := DefaultWakuConfig
	node4Config.Relay = true
	node4Config.Discv5BootstrapNodes = []string{enrNode3.String()}

	Debug("Creating Node4 with Node3 as Discv5 bootstrap")
	node4, err := StartWakuNode("Node4", &node4Config)
	require.NoError(t, err, "Failed to start Node4")

	defer func() {
		Debug("Stopping and destroying all Waku nodes")
		node1.StopAndDestroy()
		node2.StopAndDestroy()
		node3.StopAndDestroy()
		node4.StopAndDestroy()
	}()

	Debug("Waiting for nodes to auto-connect via Discv5")
	err = WaitForAutoConnection([]*WakuNode{node1, node2, node3, node4})
	require.NoError(t, err, "Nodes did not auto-connect within timeout")

	Debug("Fetching number of peers in mesh for Node1")
	peerCount, err := node1.GetNumConnectedPeers()
	require.NoError(t, err, "Failed to get number of peers in mesh for Node1")

	Debug("Total number of peers in mesh for Node1: %d", peerCount)
	require.Equal(t, 3, peerCount, "Expected Node1 to have exactly 3 peers in the mesh")

	Debug("Test successfully verified peer count in mesh with 4 nodes using Discv5 (Chained Connection)")
}
