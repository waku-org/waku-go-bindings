package waku

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBasicWakuNodes(t *testing.T) {
	Debug("Starting TestBasicWakuNodes")

	nodeCfg := DefaultWakuConfig
	nodeCfg.Relay = true

	Debug("Starting the WakuNode")
	node, err := StartWakuNode("node", &nodeCfg)
	require.NoError(t, err, "Failed to create the WakuNode")

	// Use defer to ensure proper cleanup
	defer func() {
		Debug("Stopping and destroying Node")
		node.StopAndDestroy()
	}()

	Debug("Successfully created the WakuNode")
	time.Sleep(2 * time.Second)

	Debug("TestBasicWakuNodes completed successfully")
}

func TestNodeRestart(t *testing.T) {
	Debug("Starting TestNodeRestart")

	Debug("Creating Node")
	nodeConfig := DefaultWakuConfig
	node, err := StartWakuNode("TestNode", &nodeConfig)
	require.NoError(t, err)
	Debug("Node started successfully")

	Debug("Fetching ENR before stopping the node")
	enrBefore := node.GetENR()
	require.NotEmpty(t, enrBefore)
	Debug("ENR before stopping: %s", enrBefore)

	Debug("Stopping the Node")
	err = node.Stop()
	require.NoError(t, err)
	Debug("Node stopped successfully")

	Debug("Restarting the Node")
	err = node.Start()
	require.NoError(t, err)
	Debug("Node restarted successfully")

	Debug("Fetching ENR after restarting the node")
	enrAfter := node.GetENR()
	require.NotEmpty(t, enrAfter)
	Debug("ENR after restarting: %s", enrAfter)

	Debug("Comparing ENRs before and after restart")
	require.Equal(t, enrBefore, enrAfter, "ENR should remain the same after node restart")

	Debug("Cleaning up: stopping and destroying the node")
	defer node.StopAndDestroy()

	Debug("TestNodeRestart completed successfully")
}
