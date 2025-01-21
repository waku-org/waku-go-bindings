package waku_go_tests

import (
	"testing"
    utilities "github.com/waku-org/waku-go-bindings/testlibs/utilities"
	testlibs "github.com/waku-org/waku-go-bindings/testlibs/src"
	"go.uber.org/zap"
	 "github.com/stretchr/testify/require"
)

func TestBasicWakuNodes(t *testing.T) {
     
	utilities.Debug("Create logger isntance")
	logger, err := zap.NewDevelopment()
	defaultConfig := *utilities.DefaultWakuConfig 

	
	utilities.Debug("Creating the first WakuNodeWrapper")
	node1, err := testlibs.Wrappers_CreateWakuNode(&defaultConfig, logger)
	require.NoError(t, err, "Failed to create the first WakuNodeWrapper")
	utilities.Debug("Successfully created the first WakuNodeWrapper")


	utilities.Debug("Creating the second WakuNodeWrapper")
	node2, err := testlibs.Wrappers_CreateWakuNode(&defaultConfig, logger)
	require.NoError(t, err, "Failed to create the second WakuNodeWrapper")
	utilities.Debug("Successfully created the second WakuNodeWrapper")

	
	utilities.Debug("Starting the first WakuNodeWrapper")
	err = 
	require.NoError(t, err, "Failed to start the first WakuNodeWrapper")
	utilities.Debug("Successfully started the first WakuNodeWrapper")

	
	utilities.Debug("Starting the second WakuNodeWrapper")
	err = node2.Wrappers_Start()
	require.NoError(t, err, "Failed to start the second WakuNodeWrapper")
	utilities.Debug("Successfully started the second WakuNodeWrapper")

	
	utilities.Debug("Stopping the first WakuNodeWrapper")
	err = node1.Wrappers_Stop()
	require.NoError(t, err, "Failed to stop the first WakuNodeWrapper")
	utilities.Debug("Successfully stopped the first WakuNodeWrapper")

	
	utilities.Debug("Stopping the second WakuNodeWrapper")
	err = node2.Wrappers_Stop()
	require.NoError(t, err, "Failed to stop the second WakuNodeWrapper")
	utilities.Debug("Successfully stopped the second WakuNodeWrapper")

	
	utilities.Debug("Destroying the first WakuNodeWrapper")
	err = node1.Wrappers_Destroy()
	require.NoError(t, err, "Failed to destroy the first WakuNodeWrapper")
	utilities.Debug("Successfully destroyed the first WakuNodeWrapper")


	utilities.Debug("Destroying the second WakuNodeWrapper")
	err = node2.Wrappers_Destroy()
	require.NoError(t, err, "Failed to destroy the second WakuNodeWrapper")
	utilities.Debug("Successfully destroyed the second WakuNodeWrapper")
}
