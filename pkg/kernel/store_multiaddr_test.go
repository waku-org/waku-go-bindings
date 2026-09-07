package kernel

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/require"
)

// A peer.AddrInfo carries every address libp2p knows for a peer, and all of
// them are sent. Before logos-delivery#4202 the library parsed that argument as
// a single multiaddress, so a peer with more than one address failed outright.
func TestStoreQueryWithSeveralPeerAddresses(t *testing.T) {
	storeConfig := DefaultWakuConfig
	storeConfig.Relay = true
	storeConfig.Store = true

	storeNode, err := StartWakuNode("StoreNode", &storeConfig)
	require.NoError(t, err, "Failed to start StoreNode")
	defer func() { _ = storeNode.StopAndDestroy() }()

	clientConfig := DefaultWakuConfig
	clientConfig.Relay = true

	client, err := StartWakuNode("Client", &clientConfig)
	require.NoError(t, err, "Failed to start Client")
	defer func() { _ = client.StopAndDestroy() }()

	require.NoError(t, client.ConnectPeer(storeNode))

	storeAddrs, err := storeNode.ListenAddresses()
	require.NoError(t, err)
	require.NotEmpty(t, storeAddrs)

	storeInfo, err := peer.AddrInfoFromString(storeAddrs[0].String())
	require.NoError(t, err)

	storeInfo.Addrs = append([]multiaddr.Multiaddr{
		multiaddr.StringCast("/ip4/127.0.0.1/tcp/1"),
	}, storeInfo.Addrs...)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	_, err = client.StoreQuery(ctx, &DefaultStoreQueryRequest, *storeInfo)
	require.NoError(t, err, "store query with several addresses for one peer")
}
