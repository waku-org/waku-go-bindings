package waku_go_tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	testlibs "github.com/waku-org/waku-go-bindings/testlibs/src"
	utilities "github.com/waku-org/waku-go-bindings/testlibs/utilities"
	"go.uber.org/zap"
)

func TestBasicWakuNodes(t *testing.T) {

	utilities.Debug("Create logger isntance")
	logger, _ := zap.NewDevelopment()
	nodeCfg1 := *utilities.DefaultWakuConfig
	nodeCfg1.Relay = true

	nodeCfg2 := *utilities.DefaultWakuConfig
	nodeCfg2.Relay = true

	utilities.Debug("Starting the first WakuNodeWrapper")
	node1, err := testlibs.Wrappers_StartWakuNode(&nodeCfg1, logger.Named("node1"))
	require.NoError(t, err, "Failed to create the first WakuNodeWrapper")
	utilities.Debug("Successfully created the first WakuNodeWrapper")

	utilities.Debug("Starting the second WakuNodeWrapper")
	node2, err := testlibs.Wrappers_StartWakuNode(&nodeCfg2, logger.Named("node2"))
	require.NoError(t, err, "Failed to create the second WakuNodeWrapper")
	utilities.Debug("Successfully created the second WakuNodeWrapper")

	require.NoError(t, err, "Failed to start the first WakuNodeWrapper")
	utilities.Debug("Successfully started the first WakuNodeWrapper")

	require.NoError(t, err, "Failed to start the second WakuNodeWrapper")
	utilities.Debug("Successfully started the second WakuNodeWrapper")

	time.Sleep(2 * time.Second)
	utilities.Debug("Stopping the first WakuNodeWrapper")

	err = node1.Wrappers_StopAndDestroy()
	require.NoError(t, err, "Failed to stop+destroy Node 1")

	err = node2.Wrappers_StopAndDestroy()
	require.NoError(t, err, "Failed to stop+destroy Node 2")
}

// Test to connect 2 nodes and disconnect them

func TestConnectAndDisconnectNodes(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err, "failed to create logger")

	nodeA, err := testlibs.Wrappers_StartWakuNode(nil, logger.Named("nodeA"))
	require.NoError(t, err, "failed to create/start Node A")
	defer nodeA.Wrappers_StopAndDestroy() // ensures cleanup

	nodeB, err := testlibs.Wrappers_StartWakuNode(nil, logger.Named("nodeB"))
	require.NoError(t, err, "failed to create/start Node B")
	defer nodeB.Wrappers_StopAndDestroy() // ensures cleanup

	nodeC, err := testlibs.Wrappers_StartWakuNode(nil, logger.Named("nodeB"))
	require.NoError(t, err, "failed to create/start Node B")
	defer nodeC.Wrappers_StopAndDestroy() // ensures cleanup

	err = nodeA.Wrappers_ConnectPeer(nodeB)
	require.NoError(t, err, "failed to connect Node A to Node B")

	time.Sleep(3 * time.Second)

	err = nodeA.Wrappers_DisconnectPeer(nodeC)
	require.NoError(t, err, "failed to disconnect Node A from Node B")
}
