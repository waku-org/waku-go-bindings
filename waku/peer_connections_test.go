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

	time.Sleep(3 * time.Second)

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

	Debug("Verifying connected peers for Node 3")
	connectedPeers, err := node3.GetConnectedPeers()
	require.NoError(t, err, "Failed to get connected peers for Node 3")
	node1PeerID, err := node1.PeerID()
	require.NoError(t, err, "Failed to get PeerID for Node 1")
	node2PeerID, err := node2.PeerID()
	require.NoError(t, err, "Failed to get PeerID for Node 2")

	require.True(t, slices.Contains(connectedPeers, node1PeerID), "Node 1 should be a peer of Node 3")
	require.True(t, slices.Contains(connectedPeers, node2PeerID), "Node 2 should be a peer of Node 3")

	Debug("Test completed successfully: multiple nodes connected to a single node and verified peers")
}
