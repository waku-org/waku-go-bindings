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
	utilities.Debug("Create logger instance")
	logger, _ := zap.NewDevelopment()

	nodeCfg := *utilities.DefaultWakuConfig
	nodeCfg.Relay = true

	utilities.Debug("Starting the WakuNodeWrapper")
	node, err := testlibs.StartWakuNode(&nodeCfg, logger.Named("node"))
	require.NoError(t, err, "Failed to create the WakuNodeWrapper")

	// Use defer to ensure proper cleanup
	defer func() {
		utilities.Debug("Stopping and destroying Node")
		node.StopAndDestroy()
	}()
	utilities.Debug("Successfully created the WakuNodeWrapper")

	time.Sleep(2 * time.Second)

	// No need for another StopAndDestroy here, defer already handles cleanup
	utilities.Debug("Test completed successfully")
}

func TestNodeRestart(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	logger.Debug("Starting TestNodeRestart")

	logger.Debug("Creating Node")
	nodeConfig := *utilities.DefaultWakuConfig
	node, err := testlibs.StartWakuNode(&nodeConfig, logger.Named("TestNode"))
	require.NoError(t, err)
	logger.Debug("Node started successfully")

	logger.Debug("Fetching ENR before stopping the node")
	enrBefore, _ := node.ENR()
	require.NotEmpty(t, enrBefore)
	//logger.Debug("ENR before stopping", zap.String("ENR", enrBefore))

	logger.Debug("Stopping the Node")
	err = node.Stop()
	require.NoError(t, err)
	logger.Debug("Node stopped successfully")

	logger.Debug("Restarting the Node")
	err = node.Start()
	require.NoError(t, err)
	logger.Debug("Node restarted successfully")

	logger.Debug("Fetching ENR after restarting the node")
	enrAfter, _ := node.ENR()
	require.NotEmpty(t, enrAfter)
	//logger.Debug("ENR after restarting", zap.String("ENR", enrAfter))

	logger.Debug("Comparing ENRs before and after restart")
	require.Equal(t, enrBefore, enrAfter, "ENR should remain the same after node restart")

	logger.Debug("Cleaning up: stopping and destroying the node")
	defer node.StopAndDestroy()

	logger.Debug("TestNodeRestart completed successfully")
}
