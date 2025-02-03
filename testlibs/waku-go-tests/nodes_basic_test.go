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
