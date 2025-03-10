package waku

import (
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Test node connect & disconnect peers
func TestDisconnectPeerNodes(t *testing.T) {
	Debug("Starting TestDisconnectPeerNodes")

	nodeA, err := StartWakuNode("nodeA", nil)
	require.NoError(t, err, "Failed to start Node A")
	defer nodeA.StopAndDestroy()

	nodeB, err := StartWakuNode("nodeB", nil)
	require.NoError(t, err, "Failed to start Node B")
	defer nodeB.StopAndDestroy()

	Debug("Connecting Node A to Node B")
	err = nodeA.ConnectPeer(nodeB)
	require.NoError(t, err, "Failed to connect nodes")

	Debug("Verifying connection between Node A and Node B")
	connectedPeers, err := nodeA.GetConnectedPeers()
	require.NoError(t, err, "Failed to get connected peers for Node A")
	nodeBPeerID, err := nodeB.PeerID()
	require.NoError(t, err, "Failed to get PeerID for Node B")
	require.True(t, slices.Contains(connectedPeers, nodeBPeerID), "Node B should be a peer of Node A before disconnection")

	Debug("Disconnecting Node A from Node B")
	err = nodeA.DisconnectPeer(nodeB)
	require.NoError(t, err, "Failed to disconnect nodes")

	Debug("Verifying disconnection between Node A and Node B")
	connectedPeers, err = nodeA.GetConnectedPeers()
	require.NoError(t, err, "Failed to get connected peers for Node A after disconnection")
	require.False(t, slices.Contains(connectedPeers, nodeBPeerID), "Node B should no longer be a peer of Node A after disconnection")
	Debug("Test completed successfully: Node B was disconnected from Node A")
}

func TestConnectMultipleNodesToSingleNode(t *testing.T) {
	Debug("Starting TestConnectMultipleNodesToSingleNode")

	Debug("Creating 3 nodes with automatically assigned ports")

	node1, err := StartWakuNode("node1", nil)
	require.NoError(t, err, "Failed to start Node 1")
	defer func() {
		Debug("Stopping and destroying Node 1")
		node1.StopAndDestroy()
	}()

	node2, err := StartWakuNode("node2", nil)
	require.NoError(t, err, "Failed to start Node 2")
	defer func() {
		Debug("Stopping and destroying Node 2")
		node2.StopAndDestroy()
	}()

	node3, err := StartWakuNode("node3", nil)
	require.NoError(t, err, "Failed to start Node 3")
	defer func() {
		Debug("Stopping and destroying Node 3")
		node3.StopAndDestroy()
	}()

	Debug("Connecting Node 2 to Node 1")
	err = node2.ConnectPeer(node1)
	require.NoError(t, err, "Failed to connect Node 2 to Node 1")

	Debug("Connecting Node 3 to Node 1")
	err = node3.ConnectPeer(node1)
	require.NoError(t, err, "Failed to connect Node 3 to Node 1")

	Debug("Verifying connected peers for Node 1")
	connectedPeers, err := node1.GetConnectedPeers()
	require.NoError(t, err, "Failed to get connected peers for Node 1")
	node3PeerID, err := node3.PeerID()
	require.NoError(t, err, "Failed to get PeerID for Node 1")
	node2PeerID, err := node2.PeerID()
	require.NoError(t, err, "Failed to get PeerID for Node 2")

	require.True(t, slices.Contains(connectedPeers, node3PeerID), "Node 3 should be a peer of Node 1")
	require.True(t, slices.Contains(connectedPeers, node2PeerID), "Node 2 should be a peer of Node 1")

	Debug("Test completed successfully: multiple nodes connected to a single node and verified peers")
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
	require.Equal(t, 2, peerCountBefore, "Expected Node1 to have exactly 2 peers in the mesh before stopping Node3")

	Debug("Stopping Node3")
	node3.StopAndDestroy()

	Debug("Waiting for network update after Node3 stops")
	time.Sleep(10 * time.Second)

	Debug("Fetching number of peers in mesh for Node1 after stopping Node3")
	peerCountAfter, err := node1.GetNumPeersInMesh(defaultPubsubTopic)
	require.NoError(t, err, "Failed to get number of peers in mesh for Node1 after stopping Node3")

	Debug("Total number of peers in mesh for Node1 after stopping Node3: %d", peerCountAfter)
	require.Equal(t, 1, peerCountAfter, "Expected Node1 to have exactly 1 peer in the mesh after stopping Node3")

	Debug("Test successfully verified peer count change after stopping Node3")
}

// this test commented as it will fail will be changed to have external ip in future task 
/*
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

	Debug("Fetching number of peers in connected to  Node1")
	peerCount, err := node1.GetNumConnectedPeers()
	require.NoError(t, err, "Failed to get number of peers in mesh for Node1")

	Debug("Total number of peers connected to Node1: %d", peerCount)
	require.Equal(t, 3, peerCount, "Expected Node1 to have exactly 3 peers in the mesh")

	Debug("Test successfully verified peer count in mesh with 4 nodes using Discv5 (Chained Connection)")
}
*/