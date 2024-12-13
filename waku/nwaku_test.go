package waku

import (
	"context"
	"errors"
	"slices"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	"github.com/cenkalti/backoff/v3"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/require"
	"github.com/waku-org/go-waku/waku/v2/protocol/pb"
	"github.com/waku-org/go-waku/waku/v2/protocol/store"
)

// In order to run this test, you must run an nwaku node
//
// Using Docker:
//
//	IP_ADDRESS=$(hostname -I | awk '{print $1}');
// 	docker run \
// 	-p 61000:61000/tcp -p 8000:8000/udp -p 8646:8646/tcp harbor.status.im/wakuorg/nwaku:v0.33.0 \
// 	--discv5-discovery=true --cluster-id=16 --log-level=DEBUG --shard=64 --tcp-port=61000 \
// 	--nat=extip:${IP_ADDRESS} --discv5-udp-port=8000 --rest-address=0.0.0.0 --store --rest-port=8646

func TestBasicWaku(t *testing.T) {
	extNodeRestPort := 8646
	storeNodeInfo, err := GetNwakuInfo(nil, &extNodeRestPort)
	require.NoError(t, err)

	// ctx := context.Background()

	nwakuConfig := WakuConfig{
		Nodekey:         "11d0dcea28e86f81937a3bd1163473c7fbc0a0db54fd72914849bc47bdf78710",
		Relay:           true,
		LogLevel:        "DEBUG",
		DnsDiscoveryUrl: "enrtree://AMOJVZX4V6EXP7NTJPMAYJYST2QP6AJXYW76IU6VGJS7UVSNDYZG4@boot.prod.status.nodes.status.im",
		DnsDiscovery:    true,
		Discv5Discovery: true,
		Staticnodes:     []string{storeNodeInfo.ListenAddresses[0]},
		ClusterID:       16,
		Shards:          []uint16{64},
	}

	storeNodeMa, err := ma.NewMultiaddr(storeNodeInfo.ListenAddresses[0])
	require.NoError(t, err)

	w, err := New(&nwakuConfig, nil)
	require.NoError(t, err)
	require.NoError(t, w.Start())

	enr, err := w.node.ENR()
	require.NoError(t, err)
	require.NotNil(t, enr)

	options := func(b *backoff.ExponentialBackOff) {
		b.MaxElapsedTime = 30 * time.Second
	}

	// Sanity check, not great, but it's probably helpful
	err = RetryWithBackOff(func() error {
		numConnected, err := w.node.GetNumConnectedPeers()
		if err != nil {
			return err
		}
		// Have to be connected to at least 3 nodes: the static node, the bootstrap node, and one discovered node
		if numConnected > 2 {
			return nil
		}
		return errors.New("no peers discovered")
	}, options)
	require.NoError(t, err)

	// Get local store node address
	storeNode, err := peer.AddrInfoFromString(storeNodeInfo.ListenAddresses[0])
	require.NoError(t, err)

	/*
		w.node.DialPeer(ctx, storeNode.Addrs[0], "")

		w.StorenodeCycle.SetStorenodeConfigProvider(newTestStorenodeConfigProvider(*storeNode))
	*/

	// Check that we are indeed connected to the store node
	connectedStoreNodes, err := w.node.GetPeerIDsByProtocol(store.StoreQueryID_v300)
	require.NoError(t, err)
	require.True(t, slices.Contains(connectedStoreNodes, storeNode.ID), "nwaku should be connected to the store node")

	// Disconnect from the store node
	err = w.node.DisconnectPeerByID(storeNode.ID)
	require.NoError(t, err)

	// Check that we are indeed disconnected
	connectedStoreNodes, err = w.node.GetPeerIDsByProtocol(store.StoreQueryID_v300)
	require.NoError(t, err)
	isDisconnected := !slices.Contains(connectedStoreNodes, storeNode.ID)
	require.True(t, isDisconnected, "nwaku should be disconnected from the store node")

	// Re-connect
	err = w.DialPeer(storeNodeMa)
	require.NoError(t, err)

	// Check that we are connected again
	connectedStoreNodes, err = w.node.GetPeerIDsByProtocol(store.StoreQueryID_v300)
	require.NoError(t, err)
	require.True(t, slices.Contains(connectedStoreNodes, storeNode.ID), "nwaku should be connected to the store node")

	/* filter := &common.Filter{
		PubsubTopic:   w.cfg.DefaultShardPubsubTopic,
		Messages:      common.NewMemoryMessageStore(),
		ContentTopics: common.NewTopicSetFromBytes([][]byte{{1, 2, 3, 4}}),
	}

	_, err = w.Subscribe(filter)
	require.NoError(t, err)

	msgTimestamp := w.timestamp()
	contentTopic := maps.Keys(filter.ContentTopics)[0]

	time.Sleep(2 * time.Second)

	msgID, err := w.Send(w.cfg.DefaultShardPubsubTopic, &pb.WakuMessage{
		Payload:      []byte{1, 2, 3, 4, 5},
		ContentTopic: contentTopic.ContentTopic(),
		Version:      proto.Uint32(0),
		Timestamp:    &msgTimestamp,
	}, nil)

	require.NoError(t, err)
	require.NotEqual(t, msgID, "1")

	time.Sleep(1 * time.Second)

	messages := filter.Retrieve()
	require.Len(t, messages, 1)

	timestampInSeconds := msgTimestamp / int64(time.Second)
	marginInSeconds := 20

	options = func(b *backoff.ExponentialBackOff) {
		b.MaxElapsedTime = 60 * time.Second
		b.InitialInterval = 500 * time.Millisecond
	}
	err = RetryWithBackOff(func() error {
		err := w.HistoryRetriever.Query(
			context.Background(),
			store.FilterCriteria{
				ContentFilter: protocol.NewContentFilter(w.cfg.DefaultShardPubsubTopic, contentTopic.ContentTopic()),
				TimeStart:     proto.Int64((timestampInSeconds - int64(marginInSeconds)) * int64(time.Second)),
				TimeEnd:       proto.Int64((timestampInSeconds + int64(marginInSeconds)) * int64(time.Second)),
			},
			*storeNode,
			10,
			nil, false,
		)

		return err

		// TODO-nwaku
		if err != nil || envelopeCount == 0 {
			// in case of failure extend timestamp margin up to 40secs
			if marginInSeconds < 40 {
				marginInSeconds += 5
			}
			return errors.New("no messages received from store node")
		}
		return nil

	}, options)
	require.NoError(t, err)

	time.Sleep(10 * time.Second)
	*/

	require.NoError(t, w.Stop())
}

func TestPeerExchange(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	// start node that will be discovered by PeerExchange
	discV5NodeWakuConfig := WakuConfig{
		Relay:           true,
		LogLevel:        "DEBUG",
		Discv5Discovery: true,
		ClusterID:       16,
		Shards:          []uint16{64},
		PeerExchange:    false,
		Discv5UdpPort:   9010,
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
		Relay:                true,
		LogLevel:             "DEBUG",
		Discv5Discovery:      true,
		ClusterID:            16,
		Shards:               []uint16{64},
		PeerExchange:         true,
		Discv5UdpPort:        9011,
		Discv5BootstrapNodes: []string{discv5NodeEnr.String()},
		TcpPort:              60011,
	}

	pxServerNode, err := New(&pxServerWakuConfig, logger.Named("pxServerNode"))
	require.NoError(t, err)
	require.NoError(t, pxServerNode.Start())

	// Adding an extra second to make sure PX cache is not empty
	time.Sleep(2 * time.Second)

	serverNodeMa, err := pxServerNode.ListenAddresses()
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
		Relay:            false,
		LogLevel:         "DEBUG",
		Discv5Discovery:  false,
		ClusterID:        16,
		Shards:           []uint16{64},
		PeerExchange:     true,
		Discv5UdpPort:    9012,
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
		Relay:         true,
		LogLevel:      "DEBUG",
		ClusterID:     16,
		Shards:        []uint16{64},
		Discv5UdpPort: 9020,
		TcpPort:       60020,
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
		Relay:           true,
		LogLevel:        "DEBUG",
		Discv5Discovery: false,
		ClusterID:       16,
		Shards:          []uint16{64},
		Discv5UdpPort:   9030,
		TcpPort:         60030,
	}

	dialerNode, err := New(&dialerNodeWakuConfig, logger.Named("dialerNode"))
	require.NoError(t, err)
	require.NoError(t, dialerNode.Start())
	time.Sleep(1 * time.Second)

	// start node that will receive the dial
	receiverNodeWakuConfig := WakuConfig{
		Relay:           true,
		LogLevel:        "DEBUG",
		Discv5Discovery: false,
		ClusterID:       16,
		Shards:          []uint16{64},
		Discv5UdpPort:   9031,
		TcpPort:         60031,
	}
	receiverNode, err := New(&receiverNodeWakuConfig, logger.Named("receiverNode"))
	require.NoError(t, err)
	require.NoError(t, receiverNode.Start())
	time.Sleep(1 * time.Second)
	receiverMultiaddr, err := receiverNode.ListenAddresses()
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

func TestRelay(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	// start node that will send the message
	senderNodeWakuConfig := WakuConfig{
		Relay:           true,
		LogLevel:        "DEBUG",
		Discv5Discovery: false,
		ClusterID:       16,
		Shards:          []uint16{64},
		Discv5UdpPort:   9040,
		TcpPort:         60040,
	}

	senderNode, err := New(&senderNodeWakuConfig, logger.Named("senderNode"))
	require.NoError(t, err)
	require.NoError(t, senderNode.Start())
	time.Sleep(1 * time.Second)

	// start node that will receive the message
	receiverNodeWakuConfig := WakuConfig{
		Relay:           true,
		LogLevel:        "DEBUG",
		Discv5Discovery: false,
		ClusterID:       16,
		Shards:          []uint16{64},
		Discv5UdpPort:   9041,
		TcpPort:         60041,
	}
	receiverNode, err := New(&receiverNodeWakuConfig, logger.Named("receiverNode"))
	require.NoError(t, err)
	require.NoError(t, receiverNode.Start())
	time.Sleep(1 * time.Second)
	receiverMultiaddr, err := receiverNode.ListenAddresses()
	require.NoError(t, err)
	require.NotNil(t, receiverMultiaddr)

	// Dial so they become peers
	err = senderNode.DialPeer(receiverMultiaddr[0])
	require.NoError(t, err)
	time.Sleep(1 * time.Second)
	// Check that both nodes now have one connected peer
	senderPeerCount, err := senderNode.PeerCount()
	require.NoError(t, err)
	require.True(t, senderPeerCount == 1, "Dialer node should have 1 peer")
	receiverPeerCount, err := receiverNode.PeerCount()
	require.NoError(t, err)
	require.True(t, receiverPeerCount == 1, "Receiver node should have 1 peer")

	message := &pb.WakuMessage{
		Payload:      []byte{1, 2, 3, 4, 5, 6},
		ContentTopic: "test-content-topic",
		Version:      proto.Uint32(0),
		Timestamp:    proto.Int64(time.Now().Unix()),
	}
	// send message
	pubsubTopic := FormatWakuRelayTopic(senderNodeWakuConfig.ClusterID, senderNodeWakuConfig.Shards[0])
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()
	senderNode.RelayPublish(ctx, message, pubsubTopic)

	wg := sync.WaitGroup{}
	wg.Add(1)

	// Wait to receive message
	select {
	case envelope := <-receiverNode.node.MsgChan:
		require.NotNil(t, envelope, "Envelope should be received")
		require.Equal(t, message.Payload, envelope.Message().Payload, "Received payload should match")
		require.Equal(t, message.ContentTopic, envelope.Message().ContentTopic, "Content topic should match")
		wg.Done()
	case <-time.After(10 * time.Second):
		t.Fatal("Timeout: No message received within 10 seconds")
	}
	wg.Wait()

	// Stop nodes
	require.NoError(t, senderNode.Stop())
	require.NoError(t, receiverNode.Stop())
}
