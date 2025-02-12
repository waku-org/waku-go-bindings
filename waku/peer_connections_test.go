package waku

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Test node connect & disconnect peers
func TestDisconnectPeerNodes(t *testing.T) {
	Debug("Starting TestDisconnectPeerNodes")

	// Create Node A
	nodeA, err := StartWakuNode("nodeA", nil)
	require.NoError(t, err, "Failed to start Node A")
	defer nodeA.StopAndDestroy()

	// Create Node B
	nodeB, err := StartWakuNode("nodeB", nil)
	require.NoError(t, err, "Failed to start Node B")
	defer nodeB.StopAndDestroy()

	// Connect Node A to Node B
	Debug("Connecting Node A to Node B")
	err = nodeA.ConnectPeer(nodeB)
	require.NoError(t, err, "Failed to connect nodes")

	// Wait for 3 seconds
	time.Sleep(3 * time.Second)

	// Disconnect Node A from Node B
	Debug("Disconnecting Node A from Node B")
	err = nodeA.DisconnectPeer(nodeB)
	require.NoError(t, err, "Failed to disconnect nodes")
}

func TestConnectMultipleNodesToSingleNode(t *testing.T) {
	Debug("Starting TestConnectMultipleNodesToSingleNode")

	Debug("Creating 3 nodes")
	node1, err := StartWakuNode("Node1", nil)
	require.NoError(t, err, "Failed to start Node 1")
	defer func() {
		Debug("Stopping and destroying Node 1")
		node1.StopAndDestroy()
	}()

	node2, err := StartWakuNode("Node2", nil)
	require.NoError(t, err, "Failed to start Node 2")
	defer func() {
		Debug("Stopping and destroying Node 2")
		node2.StopAndDestroy()
	}()

	node3, err := StartWakuNode("Node3", nil)
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

	Debug("Test completed successfully: multiple nodes connected to a single node")
}
