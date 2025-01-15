package waku

import (
	"context"
	"errors"
	"fmt"
	"slices"
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
	"github.com/waku-org/waku-go-bindings/waku/common"
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

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

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

	w, err := NewWakuNode(&nwakuConfig, logger.Named("nwaku"))
	require.NoError(t, err)
	require.NoError(t, w.Start())

	enr, err := w.ENR()
	require.NoError(t, err)
	require.NotNil(t, enr)

	options := func(b *backoff.ExponentialBackOff) {
		b.MaxElapsedTime = 30 * time.Second
	}

	// Sanity check, not great, but it's probably helpful
	err = RetryWithBackOff(func() error {
		numConnected, err := w.GetNumConnectedPeers()
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
	connectedStoreNodes, err := w.GetPeerIDsByProtocol(store.StoreQueryID_v300)
	require.NoError(t, err)
	require.True(t, slices.Contains(connectedStoreNodes, storeNode.ID), "nwaku should be connected to the store node")

	// Disconnect from the store node
	err = w.DisconnectPeerByID(storeNode.ID)
	require.NoError(t, err)

	// Check that we are indeed disconnected
	connectedStoreNodes, err = w.GetPeerIDsByProtocol(store.StoreQueryID_v300)
	require.NoError(t, err)
	isDisconnected := !slices.Contains(connectedStoreNodes, storeNode.ID)
	require.True(t, isDisconnected, "nwaku should be disconnected from the store node")

	// Re-connect
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()
	err = w.Connect(ctx, storeNodeMa)
	require.NoError(t, err)

	// Check that we are connected again
	connectedStoreNodes, err = w.GetPeerIDsByProtocol(store.StoreQueryID_v300)
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

	discV5Node, err := NewWakuNode(&discV5NodeWakuConfig, logger.Named("discV5Node"))
	require.NoError(t, err)
	require.NoError(t, discV5Node.Start())

	discV5NodePeerId, err := discV5Node.PeerID()
	require.NoError(t, err)

	discv5NodeEnr, err := discV5Node.ENR()
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

	pxServerNode, err := NewWakuNode(&pxServerWakuConfig, logger.Named("pxServerNode"))
	require.NoError(t, err)
	require.NoError(t, pxServerNode.Start())

	// Adding an extra second to make sure PX cache is not empty
	time.Sleep(2 * time.Second)

	serverNodeMa, err := pxServerNode.ListenAddresses()
	require.NoError(t, err)
	require.NotNil(t, serverNodeMa)
	require.True(t, len(serverNodeMa) > 0)

	// Sanity check, not great, but it's probably helpful
	options := func(b *backoff.ExponentialBackOff) {
		b.MaxElapsedTime = 30 * time.Second
	}

	// Check that pxServerNode has discV5Node in its Peer Store
	err = RetryWithBackOff(func() error {
		peers, err := pxServerNode.GetPeerIDsFromPeerStore()

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

	lightNode, err := NewWakuNode(&pxClientWakuConfig, logger.Named("lightNode"))
	require.NoError(t, err)
	require.NoError(t, lightNode.Start())

	pxServerPeerId, err := pxServerNode.PeerID()
	require.NoError(t, err)

	// Check that the light node discovered the discV5Node and has both nodes in its peer store
	err = RetryWithBackOff(func() error {
		peers, err := lightNode.GetPeerIDsFromPeerStore()
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
		numPeersReceived, err := lightNode.PeerExchangeRequest(1)
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

	node, err := NewWakuNode(&nodeWakuConfig, logger.Named("node"))
	require.NoError(t, err)
	require.NoError(t, node.Start())
	sampleEnrTree := "enrtree://AMOJVZX4V6EXP7NTJPMAYJYST2QP6AJXYW76IU6VGJS7UVSNDYZG4@boot.prod.status.nodes.status.im"

	ctx, cancel := context.WithTimeout(context.TODO(), requestTimeout)
	defer cancel()
	res, err := node.DnsDiscovery(ctx, sampleEnrTree, nameserver)
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

	dialerNode, err := NewWakuNode(&dialerNodeWakuConfig, logger.Named("dialerNode"))
	require.NoError(t, err)
	require.NoError(t, dialerNode.Start())

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

	receiverNode, err := NewWakuNode(&receiverNodeWakuConfig, logger.Named("receiverNode"))
	require.NoError(t, err)
	require.NoError(t, receiverNode.Start())
	receiverMultiaddr, err := receiverNode.ListenAddresses()
	require.NoError(t, err)
	require.NotNil(t, receiverMultiaddr)
	require.True(t, len(receiverMultiaddr) > 0)
	// Check that both nodes start with no connected peers
	dialerPeerCount, err := dialerNode.GetNumConnectedPeers()
	require.NoError(t, err)
	require.True(t, dialerPeerCount == 0, "Dialer node should have no connected peers")
	receiverPeerCount, err := receiverNode.GetNumConnectedPeers()
	require.NoError(t, err)
	require.True(t, receiverPeerCount == 0, "Receiver node should have no connected peers")
	// Dial
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()
	err = dialerNode.Connect(ctx, receiverMultiaddr[0])
	require.NoError(t, err)
	time.Sleep(1 * time.Second)
	// Check that both nodes now have one connected peer
	dialerPeerCount, err = dialerNode.GetNumConnectedPeers()
	require.NoError(t, err)
	require.True(t, dialerPeerCount == 1, "Dialer node should have 1 peer")
	receiverPeerCount, err = receiverNode.GetNumConnectedPeers()
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

	senderNode, err := NewWakuNode(&senderNodeWakuConfig, logger.Named("senderNode"))
	require.NoError(t, err)
	require.NoError(t, senderNode.Start())

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
	receiverNode, err := NewWakuNode(&receiverNodeWakuConfig, logger.Named("receiverNode"))
	require.NoError(t, err)
	require.NoError(t, receiverNode.Start())
	receiverMultiaddr, err := receiverNode.ListenAddresses()
	require.NoError(t, err)
	require.NotNil(t, receiverMultiaddr)
	require.True(t, len(receiverMultiaddr) > 0)

	// Dial so they become peers
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()
	err = senderNode.Connect(ctx, receiverMultiaddr[0])
	require.NoError(t, err)
	time.Sleep(1 * time.Second)
	// Check that both nodes now have one connected peer
	senderPeerCount, err := senderNode.GetNumConnectedPeers()
	require.NoError(t, err)
	require.True(t, senderPeerCount == 1, "Dialer node should have 1 peer")
	receiverPeerCount, err := receiverNode.GetNumConnectedPeers()
	require.NoError(t, err)
	require.True(t, receiverPeerCount == 1, "Receiver node should have 1 peer")

	message := &pb.WakuMessage{
		Payload:      []byte{1, 2, 3, 4, 5, 6},
		ContentTopic: "test-content-topic",
		Version:      proto.Uint32(0),
		Timestamp:    proto.Int64(time.Now().UnixNano()),
	}
	// send message
	pubsubTopic := FormatWakuRelayTopic(senderNodeWakuConfig.ClusterID, senderNodeWakuConfig.Shards[0])
	ctx2, cancel2 := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel2()
	senderNode.RelayPublish(ctx2, message, pubsubTopic)

	// Wait to receive message
	select {
	case envelope := <-receiverNode.MsgChan:
		require.NotNil(t, envelope, "Envelope should be received")
		require.Equal(t, message.Payload, envelope.Message().Payload, "Received payload should match")
		require.Equal(t, message.ContentTopic, envelope.Message().ContentTopic, "Content topic should match")
	case <-time.After(10 * time.Second):
		t.Fatal("Timeout: No message received within 10 seconds")
	}

	// Stop nodes
	require.NoError(t, senderNode.Stop())
	require.NoError(t, receiverNode.Stop())
}

func TestTopicHealth(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	clusterId := uint16(16)
	shardId := uint16(64)

	// start node1
	wakuConfig1 := WakuConfig{
		Relay:           true,
		LogLevel:        "DEBUG",
		Discv5Discovery: false,
		ClusterID:       clusterId,
		Shards:          []uint16{shardId},
		Discv5UdpPort:   9050,
		TcpPort:         60050,
	}

	node1, err := NewWakuNode(&wakuConfig1, logger.Named("node1"))
	require.NoError(t, err)
	require.NoError(t, node1.Start())

	// start node2
	wakuConfig2 := WakuConfig{
		Relay:           true,
		LogLevel:        "DEBUG",
		Discv5Discovery: false,
		ClusterID:       clusterId,
		Shards:          []uint16{shardId},
		Discv5UdpPort:   9051,
		TcpPort:         60051,
	}
	node2, err := NewWakuNode(&wakuConfig2, logger.Named("node2"))
	require.NoError(t, err)
	require.NoError(t, node2.Start())
	multiaddr2, err := node2.ListenAddresses()
	require.NoError(t, err)
	require.NotNil(t, multiaddr2)
	require.True(t, len(multiaddr2) > 0)

	// node1 dials node2 so they become peers
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()
	err = node1.Connect(ctx, multiaddr2[0])
	require.NoError(t, err)
	time.Sleep(1 * time.Second)
	// Check that both nodes now have one connected peer
	peerCount1, err := node1.GetNumConnectedPeers()
	require.NoError(t, err)
	require.True(t, peerCount1 == 1, "node1 should have 1 peer")
	peerCount2, err := node2.GetNumConnectedPeers()
	require.NoError(t, err)
	require.True(t, peerCount2 == 1, "node2 should have 1 peer")

	// Wait to receive topic health update
	select {
	case topicHealth := <-node2.TopicHealthChan:
		require.NotNil(t, topicHealth, "topicHealth should be updated")
		require.Equal(t, topicHealth.TopicHealth, "MinimallyHealthy", "Topic health should be MinimallyHealthy")
		require.Equal(t, topicHealth.PubsubTopic, FormatWakuRelayTopic(clusterId, shardId), "PubsubTopic should match configured cluster and shard")
	case <-time.After(10 * time.Second):
		t.Fatal("Timeout: No topic health event received within 10 seconds")
	}

	// Stop nodes
	require.NoError(t, node1.Stop())
	require.NoError(t, node2.Stop())

}

func TestConnectionChange(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	clusterId := uint16(16)
	shardId := uint16(64)

	// start node1
	wakuConfig1 := WakuConfig{
		Relay:           true,
		LogLevel:        "DEBUG",
		Discv5Discovery: false,
		ClusterID:       clusterId,
		Shards:          []uint16{shardId},
		Discv5UdpPort:   9060,
		TcpPort:         60060,
	}

	node1, err := NewWakuNode(&wakuConfig1, logger.Named("node1"))
	require.NoError(t, err)
	require.NoError(t, node1.Start())

	// start node2
	wakuConfig2 := WakuConfig{
		Relay:           true,
		LogLevel:        "DEBUG",
		Discv5Discovery: false,
		ClusterID:       clusterId,
		Shards:          []uint16{shardId},
		Discv5UdpPort:   9061,
		TcpPort:         60061,
	}
	node2, err := NewWakuNode(&wakuConfig2, logger.Named("node2"))
	require.NoError(t, err)
	require.NoError(t, node2.Start())
	multiaddr2, err := node2.ListenAddresses()
	require.NoError(t, err)
	require.NotNil(t, multiaddr2)
	require.True(t, len(multiaddr2) > 0)

	// node1 dials node2 so they become peers
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()
	err = node1.Connect(ctx, multiaddr2[0])
	require.NoError(t, err)
	time.Sleep(1 * time.Second)
	// Check that both nodes now have one connected peer
	peerCount1, err := node1.GetNumConnectedPeers()
	require.NoError(t, err)
	require.True(t, peerCount1 == 1, "node1 should have 1 peer")
	peerCount2, err := node2.GetNumConnectedPeers()
	require.NoError(t, err)
	require.True(t, peerCount2 == 1, "node2 should have 1 peer")

	peerId1, err := node1.PeerID()
	require.NoError(t, err)

	// Wait to receive connectionChange event
	select {
	case connectionChange := <-node2.ConnectionChangeChan:
		require.NotNil(t, connectionChange, "connectionChange should be updated")
		require.Equal(t, connectionChange.PeerEvent, "Joined", "connectionChange Joined event should be emitted")
		require.Equal(t, connectionChange.PeerId, peerId1, "connectionChange event should contain node 1's peerId")
	case <-time.After(10 * time.Second):
		t.Fatal("Timeout: No connectionChange event received within 10 seconds")
	}

	// Disconnect from node1
	err = node2.DisconnectPeerByID(peerId1)
	require.NoError(t, err)

	// Wait to receive connectionChange event
	select {
	case connectionChange := <-node2.ConnectionChangeChan:
		require.NotNil(t, connectionChange, "connectionChange should be updated")
		require.Equal(t, connectionChange.PeerEvent, "Left", "connectionChange Left event should be emitted")
		require.Equal(t, connectionChange.PeerId, peerId1, "connectionChange event should contain node 1's peerId")
	case <-time.After(10 * time.Second):
		t.Fatal("Timeout: No connectionChange event received within 10 seconds")
	}

	// Stop nodes
	require.NoError(t, node1.Stop())
	require.NoError(t, node2.Stop())
}

func TestHash(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	// start node that will send the message
	senderNodeWakuConfig := WakuConfig{
		Relay:           true,
		Store:           true,
		LogLevel:        "DEBUG",
		Discv5Discovery: false,
		ClusterID:       16,
		Shards:          []uint16{64},
		Discv5UdpPort:   9070,
		TcpPort:         60070,
		LegacyStore:     false,
	}

	fmt.Println("------------ creating node 1")
	senderNode, err := NewWakuNode(&senderNodeWakuConfig, logger.Named("senderNode"))
	require.NoError(t, err)
	fmt.Println("------------ starting node 1")
	require.NoError(t, senderNode.Start())
	fmt.Println("------------ node 1 started")

	// start node that will receive the message
	receiverNodeWakuConfig := WakuConfig{
		Relay:           true,
		Store:           true,
		LogLevel:        "DEBUG",
		Discv5Discovery: false,
		ClusterID:       16,
		Shards:          []uint16{64},
		Discv5UdpPort:   9071,
		TcpPort:         60071,
		LegacyStore:     false,
	}
	receiverNode, err := NewWakuNode(&receiverNodeWakuConfig, logger.Named("receiverNode"))
	require.NoError(t, err)
	require.NoError(t, receiverNode.Start())
	receiverMultiaddr, err := receiverNode.ListenAddresses()
	require.NoError(t, err)
	require.NotNil(t, receiverMultiaddr)
	require.True(t, len(receiverMultiaddr) > 0)

	// Dial so they become peers
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()
	err = senderNode.Connect(ctx, receiverMultiaddr[0])
	require.NoError(t, err)
	time.Sleep(1 * time.Second)
	// Check that both nodes now have one connected peer
	senderPeerCount, err := senderNode.GetNumConnectedPeers()
	require.NoError(t, err)
	require.True(t, senderPeerCount == 1, "Dialer node should have 1 peer")
	receiverPeerCount, err := receiverNode.GetNumConnectedPeers()
	require.NoError(t, err)
	require.True(t, receiverPeerCount == 1, "Receiver node should have 1 peer")

	message := &pb.WakuMessage{
		Payload:      []byte{1, 2, 3, 4, 5, 6},
		ContentTopic: "test-content-topic",
		Version:      proto.Uint32(0),
		Timestamp:    proto.Int64(time.Now().UnixNano()),
	}
	// send message
	pubsubTopic := FormatWakuRelayTopic(senderNodeWakuConfig.ClusterID, senderNodeWakuConfig.Shards[0])
	ctx2, cancel2 := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel2()

	fmt.Println("----------- publishing message ------------")
	hash, err := senderNode.RelayPublish(ctx2, message, pubsubTopic)
	require.NoError(t, err)
	fmt.Println("----------- RelayPublish returned hash: ", hash)

	// Wait to receive message
	select {
	case envelope := <-receiverNode.MsgChan:
		fmt.Println("------- received envelope: ", envelope)
		require.NotNil(t, envelope, "Envelope should be received")
		require.Equal(t, message.Payload, envelope.Message().Payload, "Received payload should match")
		require.Equal(t, message.ContentTopic, envelope.Message().ContentTopic, "Content topic should match")
	case <-time.After(10 * time.Second):
		t.Fatal("Timeout: No message received within 10 seconds")
	}

	// Now send store query
	ctx3, cancel3 := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel3()

	storeReq := common.StoreQueryRequest{
		IncludeData:   true,
		ContentTopics: []string{"test-content-topic"},
	}

	fmt.Println("------------ storeNode multiaddr: ", receiverMultiaddr[0].String())
	storeNodeAddrInfo, err := peer.AddrInfoFromString(receiverMultiaddr[0].String())
	require.NoError(t, err)

	res, err := senderNode.StoreQuery(ctx3, &storeReq, *storeNodeAddrInfo)
	require.NoError(t, err)
	fmt.Printf("%+v\n", res)

	// Stop nodes
	require.NoError(t, senderNode.Stop())
	require.NoError(t, receiverNode.Stop())
}
