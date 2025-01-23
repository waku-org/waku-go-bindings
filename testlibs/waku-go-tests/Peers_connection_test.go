package waku_go_tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	testlibs "github.com/waku-org/waku-go-bindings/testlibs/src"
	utilities "github.com/waku-org/waku-go-bindings/testlibs/utilities"
	"go.uber.org/zap"
)

// test  node connect & disconnect peers
func TestDisconnectPeerNodes(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	// Create Node A
	nodeA, err := testlibs.Wrappers_StartWakuNode(nil, logger.Named("nodeA"))
	require.NoError(t, err)
	defer nodeA.Wrappers_StopAndDestroy()

	// Create Node B
	nodeB, err := testlibs.Wrappers_StartWakuNode(nil, logger.Named("nodeB"))
	require.NoError(t, err)
	defer nodeB.Wrappers_StopAndDestroy()

	// Connect Node A to Node B
	err = nodeA.Wrappers_ConnectPeer(nodeB)
	require.NoError(t, err, "failed  to connect nodes")

	// Wait for 3 seconds
	time.Sleep(3 * time.Second)

	// Disconnect Node A from Node B
	err = nodeA.Wrappers_DisconnectPeer(nodeB)
	require.NoError(t, err, "failed to disconnect nodes")
}

func TestConnectMultipleNodesToSingleNode(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	utilities.Debug("Starting test to connect multiple nodes to a single node")
	utilities.Debug("Create 3 nodes")
	node1, err := testlibs.Wrappers_StartWakuNode(nil, logger.Named("Node1"))
	require.NoError(t, err)
	defer func() {
		utilities.Debug("Stopping and destroying Node 1")
		node1.Wrappers_StopAndDestroy()
	}()

	node2, err := testlibs.Wrappers_StartWakuNode(nil, logger.Named("Node2"))
	require.NoError(t, err)
	defer func() {
		utilities.Debug("Stopping and destroying Node 2")
		node2.Wrappers_StopAndDestroy()
	}()

	node3, err := testlibs.Wrappers_StartWakuNode(nil, logger.Named("Node3"))
	require.NoError(t, err)
	defer func() {
		utilities.Debug("Stopping and destroying Node 3")
		node3.Wrappers_StopAndDestroy()
	}()

	utilities.Debug("Connecting Node 2 to Node 1")
	err = node2.Wrappers_ConnectPeer(node1)
	require.NoError(t, err)

	utilities.Debug("Connecting Node 3 to Node 1")
	err = node3.Wrappers_ConnectPeer(node1)
	require.NoError(t, err)

	utilities.Debug("Test completed successfully: multiple nodes connected to a single node")
}
