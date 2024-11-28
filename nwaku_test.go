//go:build use_nwaku
// +build use_nwaku

package wakuv2

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/cenkalti/backoff/v3"
	"github.com/stretchr/testify/require"
)

func TestPeerExchange(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	// start node that will be discovered by PeerExchange
	discV5NodeWakuConfig := WakuConfig{
		EnableRelay:     true,
		LogLevel:        "DEBUG",
		Discv5Discovery: true,
		ClusterID:       16,
		Shards:          []uint16{64},
		PeerExchange:    false,
		Discv5UdpPort:   9001,
		TcpPort:         60010,
	}

	discV5Node, err := New(&discV5NodeWakuConfig, logger.Named("discV5Node"))
	require.NoError(t, err)
	require.NoError(t, discV5Node.Start())

	time.Sleep(1 * time.Second)

	discV5NodePeerId, err := discV5Node.node.PeerID()
	require.NoError(t, err)

	discv5NodeEnr, err := discV5Node.node.ENR()
	require.NoError(t, err)

	// start node which serves as PeerExchange server
	pxServerWakuConfig := WakuConfig{
		EnableRelay:          true,
		LogLevel:             "DEBUG",
		Discv5Discovery:      true,
		ClusterID:            16,
		Shards:               []uint16{64},
		PeerExchange:         true,
		Discv5UdpPort:        9000,
		Discv5BootstrapNodes: []string{discv5NodeEnr.String()},
		TcpPort:              60011,
	}

	pxServerNode, err := New(&pxServerWakuConfig, logger.Named("pxServerNode"))
	require.NoError(t, err)
	require.NoError(t, pxServerNode.Start())

	// Adding an extra second to make sure PX cache is not empty
	time.Sleep(2 * time.Second)

	serverNodeMa, err := pxServerNode.node.ListenAddresses()
	require.NoError(t, err)
	require.NotNil(t, serverNodeMa)

	// Sanity check, not great, but it's probably helpful
	options := func(b *backoff.ExponentialBackOff) {
		b.MaxElapsedTime = 30 * time.Second
	}

	// Check that pxServerNode has discV5Node in its Peer Store
	err = RetryWithBackOff(func() error {
		peers, err := pxServerNode.node.GetPeerIDsFromPeerStore()

		if err != nil {
			return err
		}

		if slices.Contains(peers, discV5NodePeerId) {
			return nil
		}

		return errors.New("pxServer is missing the discv5 node in its peer store")
	}, options)
	require.NoError(t, err)

	// start light node which uses PeerExchange to discover peers
	pxClientWakuConfig := WakuConfig{
		EnableRelay:      false,
		LogLevel:         "DEBUG",
		Discv5Discovery:  false,
		ClusterID:        16,
		Shards:           []uint16{64},
		PeerExchange:     true,
		Discv5UdpPort:    9002,
		TcpPort:          60012,
		PeerExchangeNode: serverNodeMa[0].String(),
	}

	lightNode, err := New(&pxClientWakuConfig, logger.Named("lightNode"))
	require.NoError(t, err)
	require.NoError(t, lightNode.Start())

	time.Sleep(1 * time.Second)

	pxServerPeerId, err := pxServerNode.node.PeerID()
	require.NoError(t, err)

	// Check that the light node discovered the discV5Node and has both nodes in its peer store
	err = RetryWithBackOff(func() error {
		peers, err := lightNode.node.GetPeerIDsFromPeerStore()
		if err != nil {
			return err
		}

		if slices.Contains(peers, discV5NodePeerId) && slices.Contains(peers, pxServerPeerId) {
			return nil
		}
		return errors.New("lightnode is missing peers")
	}, options)
	require.NoError(t, err)

	// Now perform the PX request manually to see if it also works
	err = RetryWithBackOff(func() error {
		numPeersReceived, err := lightNode.node.PeerExchangeRequest(1)
		if err != nil {
			return err
		}

		if numPeersReceived == 1 {
			return nil
		}
		return errors.New("Peer Exchange is not returning peers")
	}, options)
	require.NoError(t, err)

	// Stop nodes
	require.NoError(t, lightNode.Stop())
	require.NoError(t, pxServerNode.Stop())
	require.NoError(t, discV5Node.Stop())

}

func TestDnsDiscover(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	nameserver := "8.8.8.8"
	nodeWakuConfig := WakuConfig{
		EnableRelay:   true,
		LogLevel:      "DEBUG",
		ClusterID:     16,
		Shards:        []uint16{64},
		Discv5UdpPort: 9040,
		TcpPort:       60040,
	}
	node, err := New(&nodeWakuConfig, logger.Named("node"))
	require.NoError(t, err)
	require.NoError(t, node.Start())
	time.Sleep(1 * time.Second)
	sampleEnrTree := "enrtree://AMOJVZX4V6EXP7NTJPMAYJYST2QP6AJXYW76IU6VGJS7UVSNDYZG4@boot.prod.status.nodes.status.im"

	ctx, cancel := context.WithTimeout(context.TODO(), requestTimeout)
	defer cancel()
	res, err := node.node.DnsDiscovery(ctx, sampleEnrTree, nameserver)
	require.NoError(t, err)
	require.True(t, len(res) > 1, "multiple nodes should be returned from the DNS Discovery query")
	// Stop nodes
	require.NoError(t, node.Stop())
}

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

	dialerNode, err := New(&dialerNodeWakuConfig, logger.Named("dialerNode"))
	require.NoError(t, err)
	require.NoError(t, dialerNode.Start())
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
	require.NoError(t, err)
	require.NoError(t, receiverNode.Start())
	time.Sleep(1 * time.Second)
	receiverMultiaddr, err := receiverNode.node.ListenAddresses()
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
