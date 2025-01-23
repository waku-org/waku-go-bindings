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

