//go:build use_nwaku
// +build use_nwaku

package wakuv2

import (
	"fmt"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/stretchr/testify/require"
)

func TestDial(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	// start node that will initiate the dial
	dialerNodeWakuConfig := WakuConfig{
		EnableRelay:     true,
		LogLevel:        "DEBUG",
		Discv5Discovery: false,
		ClusterID:       16,
		Shards:          []uint16{64},
		Discv5UdpPort:   9020,
		TcpPort:         60020,
	}

	fmt.Println("------------ 1 -------------")
	dialerNode, err := New(&dialerNodeWakuConfig, logger.Named("dialerNode"))
	require.NoError(t, err)
	fmt.Println("------------ 2 -------------")
	require.NoError(t, dialerNode.Start())
	fmt.Println("------------ 3 -------------")
	time.Sleep(1 * time.Second)

	// start node that will receive the dial
	receiverNodeWakuConfig := WakuConfig{
		EnableRelay:     true,
		LogLevel:        "DEBUG",
		Discv5Discovery: false,
		ClusterID:       16,
		Shards:          []uint16{64},
		Discv5UdpPort:   9021,
		TcpPort:         60021,
	}
	receiverNode, err := New(&receiverNodeWakuConfig, logger.Named("receiverNode"))
	fmt.Println("------------ 4 -------------")
	require.NoError(t, err)
	require.NoError(t, receiverNode.Start())
	fmt.Println("------------ 5 -------------")
	time.Sleep(1 * time.Second)
	fmt.Println("------------ 6 -------------")
	receiverMultiaddr, err := receiverNode.node.ListenAddresses()
	fmt.Println("------------ 7 -------------")
	require.NoError(t, err)
	require.NotNil(t, receiverMultiaddr)
	// Check that both nodes start with no connected peers
	dialerPeerCount, err := dialerNode.PeerCount()
	require.NoError(t, err)
	require.True(t, dialerPeerCount == 0, "Dialer node should have no connected peers")
	receiverPeerCount, err := receiverNode.PeerCount()
	require.NoError(t, err)
	require.True(t, receiverPeerCount == 0, "Receiver node should have no connected peers")
	// Dial
	err = dialerNode.DialPeer(receiverMultiaddr[0])
	require.NoError(t, err)
	time.Sleep(1 * time.Second)
	// Check that both nodes now have one connected peer
	dialerPeerCount, err = dialerNode.PeerCount()
	require.NoError(t, err)
	require.True(t, dialerPeerCount == 1, "Dialer node should have 1 peer")
	receiverPeerCount, err = receiverNode.PeerCount()
	require.NoError(t, err)
	require.True(t, receiverPeerCount == 1, "Receiver node should have 1 peer")
	// Stop nodes
	require.NoError(t, dialerNode.Stop())
	require.NoError(t, receiverNode.Stop())
}
