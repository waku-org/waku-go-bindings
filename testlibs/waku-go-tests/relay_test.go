package waku_go_tests

import (
	"testing"

	"github.com/stretchr/testify/require"
	testlibs "github.com/waku-org/waku-go-bindings/testlibs/src"
	utilities "github.com/waku-org/waku-go-bindings/testlibs/utilities"
	"go.uber.org/zap"
)

func TestRelaySubscribeToDefaultTopic(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	utilities.Debug("Starting test to verify relay subscription to the default pubsub topic")

	// Define the configuration with relay = true
	wakuConfig := *utilities.DefaultWakuConfig
	wakuConfig.Relay = true

	utilities.Debug("Creating a Waku node with relay enabled")
	node, err := testlibs.Wrappers_StartWakuNode(&wakuConfig, logger.Named("TestNode"))
	require.NoError(t, err)
	defer func() {
		utilities.Debug("Stopping and destroying the Waku node")
		node.Wrappers_StopAndDestroy()
	}()

	defaultPubsubTopic := utilities.DefaultPubsubTopic
	utilities.Debug("Default pubsub topic retrieved", zap.String("topic", defaultPubsubTopic))

	utilities.Debug("Fetching number of connected relay peers before subscription", zap.String("topic", defaultPubsubTopic))
	numPeersBefore, err := node.Wrappers_GetNumConnectedRelayPeers(defaultPubsubTopic)
	require.NoError(t, err)
	utilities.Debug("Number of connected relay peers before subscription", zap.Int("count", numPeersBefore))

	utilities.Debug("Attempting to subscribe to the default pubsub topic", zap.String("topic", defaultPubsubTopic))
	err = node.Wrappers_RelaySubscribe(defaultPubsubTopic)
	require.NoError(t, err)

	utilities.Debug("Fetching number of connected relay peers after subscription", zap.String("topic", defaultPubsubTopic))
	numPeersAfter, err := node.Wrappers_GetNumConnectedRelayPeers(defaultPubsubTopic)
	require.NoError(t, err)
	utilities.Debug("Number of connected relay peers after subscription", zap.Int("count", numPeersAfter))

	require.Greater(t, numPeersAfter, numPeersBefore, "Number of connected relay peers should increase after subscription")

	utilities.Debug("Test successfully verified subscription to the default pubsub topic", zap.String("topic", defaultPubsubTopic))
}
