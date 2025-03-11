package waku

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/waku-org/waku-go-bindings/waku/common"
	"google.golang.org/protobuf/proto"
)

func TestMemoryUsageForThreeNodes(t *testing.T) {
	logFile, err := os.OpenFile("test_logs.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	require.NoError(t, err)
	defer logFile.Close()

	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)
	logger := _getLogger()
	logger.SetOutput(multiWriter)
	logger.SetLevel(logrus.DebugLevel)

	testName := t.Name()

	var memStats runtime.MemStats

	runtime.ReadMemStats(&memStats)
	Debug("[%s] Memory usage BEFORE creating nodes: %d KB", testName, memStats.HeapAlloc/1024)

	node1Cfg := DefaultWakuConfig
	node1Cfg.TcpPort, node1Cfg.Discv5UdpPort, err = GetFreePortIfNeeded(0, 0)
	require.NoError(t, err)
	node2Cfg := DefaultWakuConfig
	node2Cfg.TcpPort, node2Cfg.Discv5UdpPort, err = GetFreePortIfNeeded(0, 0)
	require.NoError(t, err)
	node3Cfg := DefaultWakuConfig
	node3Cfg.TcpPort, node3Cfg.Discv5UdpPort, err = GetFreePortIfNeeded(0, 0)
	require.NoError(t, err)

	node1, err := NewWakuNode(&node1Cfg, "node1")
	require.NoError(t, err)
	node2, err := NewWakuNode(&node2Cfg, "node2")
	require.NoError(t, err)
	node3, err := NewWakuNode(&node3Cfg, "node3")
	require.NoError(t, err)

	runtime.ReadMemStats(&memStats)
	Debug("[%s] Memory usage AFTER creating nodes: %d KB", testName, memStats.HeapAlloc/1024)

	err = node1.Start()
	require.NoError(t, err)
	err = node2.Start()
	require.NoError(t, err)
	err = node3.Start()
	require.NoError(t, err)

	runtime.ReadMemStats(&memStats)
	Debug("[%s] Memory usage AFTER starting nodes: %d KB", testName, memStats.HeapAlloc/1024)

	time.Sleep(2 * time.Second)

	node1.StopAndDestroy()
	node2.StopAndDestroy()
	node3.StopAndDestroy()

	runtime.GC()
	time.Sleep(1 * time.Second)
	runtime.GC()

	runtime.ReadMemStats(&memStats)
	Debug("[%s] Memory usage AFTER destroying nodes: %d KB", testName, memStats.HeapAlloc/1024)

	Debug("[%s] Test completed successfully", testName)
}

func Test5Nodes1kTearDown(t *testing.T) {
	logFile, err := os.OpenFile("test_repeated_start_stop.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	require.NoError(t, err)
	defer logFile.Close()

	multiWriter := io.MultiWriter(os.Stdout, logFile)
	logger := _getLogger()
	logger.SetOutput(multiWriter)
	logger.SetLevel(logrus.DebugLevel)

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	initialMem := memStats.HeapAlloc
	Debug("[%s] Memory usage at test START: %d KB", t.Name(), initialMem/1024)

	initialRSS, err := getRSSKB()
	require.NoError(t, err)
	Debug("[%s] OS-level RSS at test START: %d KB", t.Name(), initialRSS)

	//totalIterations := 3
	//for i := 1; i <= totalIterations; i++ {
	var nodes []*WakuNode
	cfg1 := DefaultWakuConfig
	cfg1.Relay = true
	node, err := StartWakuNode("node1", &cfg1)
	require.NoError(t, err, "Failed to start node%d", 0)
	nodes = append(nodes, node)
	cfg2 := DefaultWakuConfig
	cfg2.Relay = true
	node2, err := StartWakuNode("node2", &cfg2)
	require.NoError(t, err, "Failed to start node%d", 1)
	nodes = append(nodes, node2)
	time.Sleep(5 * time.Second)

	err = node.ConnectPeer(node2)
	require.NoError(t, err)
	err = WaitForAutoConnection(nodes)
	require.NoError(t, err)
	message := node.CreateMessage()
	msgHash, err := node.RelayPublishNoCTX(DefaultPubsubTopic, message)
	require.NoError(t, err)
	time.Sleep(5 * time.Second)
	err = node2.VerifyMessageReceived(message, msgHash)
	require.NoError(t, err, "Node5 did not receive message from node1 at iteration %d", 0)

	//node.StopAndDestroy()
	//node2.StopAndDestroy()
	//time.Sleep(20 * time.Second)
	cfg3 := DefaultWakuConfig
	cfg3.Relay = true
	node3, err := StartWakuNode("node1", &cfg3)
	require.NoError(t, err, "Failed to start node%d", 0)

	cfg4 := DefaultWakuConfig
	cfg4.Relay = true
	node4, err := StartWakuNode("node2", &cfg4)
	require.NoError(t, err, "Failed to start node%d", 1)
	//nodes = append(nodes, node2)
	time.Sleep(5 * time.Second)

	err = node3.ConnectPeer(node4)
	require.NoError(t, err)
	time.Sleep(5 * time.Second)
	require.NoError(t, err)
	message2 := node3.CreateMessage()
	msgHash2, err := node3.RelayPublishNoCTX(DefaultPubsubTopic, message2)
	require.NoError(t, err)
	time.Sleep(5 * time.Second)
	err = node4.VerifyMessageReceived(message2, msgHash2)
	require.NoError(t, err, "Node5 did not receive message from node1 at iteration %d", 1)

	// }
}

func TestSendMessagesInTwoPhases(t *testing.T) {

	node1Config := DefaultWakuConfig
	node1Config.Relay = true
	node1, err := StartWakuNode("Node1", &node1Config)
	require.NoError(t, err, "Failed to start Node1")

	node2Config := DefaultWakuConfig
	node2Config.Relay = true
	node2, err := StartWakuNode("Node2", &node2Config)
	require.NoError(t, err, "Failed to start Node2")

	err = node1.ConnectPeer(node2)
	require.NoError(t, err, "Failed to connect Node1 to Node2")

	msgFromNode1 := node1.CreateMessage()
	msgHash1, err := node1.RelayPublishNoCTX(DefaultPubsubTopic, msgFromNode1)
	require.NoError(t, err, "Failed to publish from Node1")

	time.Sleep(500 * time.Millisecond)

	err = node2.VerifyMessageReceived(msgFromNode1, msgHash1, 2*time.Second)
	require.NoError(t, err, "Node2 did not receive message from Node1")

	//node1.StopAndDestroy()
	//node2.StopAndDestroy()

	node3Config := DefaultWakuConfig
	node3Config.Relay = true
	node3, err := StartWakuNode("Node3", &node3Config)
	require.NoError(t, err, "Failed to start Node3")

	node4Config := DefaultWakuConfig
	node4Config.Relay = true
	node4, err := StartWakuNode("Node4", &node4Config)
	require.NoError(t, err, "Failed to start Node4")

	err = node3.ConnectPeer(node4)
	require.NoError(t, err, "Failed to connect Node3 to Node4")

	msgFromNode3 := node3.CreateMessage()
	msgHash3, err := node3.RelayPublishNoCTX(DefaultPubsubTopic, msgFromNode3)
	require.NoError(t, err, "Failed to publish from Node3")

	time.Sleep(500 * time.Millisecond)

	err = node4.VerifyMessageReceived(msgFromNode3, msgHash3, 2*time.Second)
	require.NoError(t, err, "Node4 did not receive message from Node3")

	node3.StopAndDestroy()
	node4.StopAndDestroy()
}

func TestStoreQuery5kMessagesWithPagination(t *testing.T) {
	Debug("Starting test")

	nodeConfig := DefaultWakuConfig
	nodeConfig.Relay = true
	nodeConfig.Store = true

	Debug("Creating 2 nodes")
	wakuNode, err := StartWakuNode("node1", &nodeConfig)
	require.NoError(t, err, "Failed to start Waku node")

	node2, err := StartWakuNode("node2", &nodeConfig)
	require.NoError(t, err, "Failed to start Waku node")
	node2.ConnectPeer(wakuNode)

	defer func() {
		Debug("Stopping and destroying Waku node")
		wakuNode.StopAndDestroy()
		node2.StopAndDestroy()
	}()

	var memStats runtime.MemStats
	iterations := 50

	runtime.ReadMemStats(&memStats)
	initialHeapAlloc := memStats.HeapAlloc
	Debug("Initial memory usage check before publishing %d MB", initialHeapAlloc/1024/1024)

	queryTimestamp := proto.Int64(time.Now().UnixNano())

	for i := 0; i < iterations; i++ {
		message := wakuNode.CreateMessage()
		message.Payload = []byte(fmt.Sprintf("Test endurance message payload %d", i))
		_, err := wakuNode.RelayPublishNoCTX(DefaultPubsubTopic, message)
		require.NoError(t, err, "Failed to publish message")

		if i%10 == 0 {
			runtime.ReadMemStats(&memStats)
			Debug("Memory usage at iteration %d: HeapAlloc=%v MB, NumGC=%v",
				i, memStats.HeapAlloc/1024/1024, memStats.NumGC)

			storeQueryRequest := &common.StoreQueryRequest{
				TimeStart:       queryTimestamp,
				IncludeData:     true,
				PaginationLimit: proto.Uint64(50),
			}

			storedmsgs, err := wakuNode.GetStoredMessages(node2, storeQueryRequest)
			require.NoError(t, err, "Failed to query store messages")
			require.Greater(t, len(*storedmsgs.Messages), 0, "Expected at least one stored message")
		}
	}

	runtime.ReadMemStats(&memStats)
	finalHeapAlloc := memStats.HeapAlloc
	Debug("Memory before test: %v KB, Memory after test: %v KB", initialHeapAlloc/1024, finalHeapAlloc/1024)

	require.LessOrEqual(t, finalHeapAlloc, initialHeapAlloc*2, "Memory usage has grown too much")

	Debug("Test completed successfully")
}
