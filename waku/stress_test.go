//go:build !stress
// +build !stress

package waku

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/waku-org/waku-go-bindings/utils"
	"github.com/waku-org/waku-go-bindings/waku/common"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/protobuf/proto"
)

func TestMemoryUsageForThreeNodes(t *testing.T) {
	logFile, err := os.OpenFile("test_logs.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	require.NoError(t, err)
	defer logFile.Close()

	multiWriter := io.MultiWriter(os.Stdout, logFile)
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "ts"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderCfg),
		zapcore.AddSync(multiWriter),
		zap.DebugLevel,
	)
	customLogger := zap.New(core)
	SetLogger(customLogger)

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

func Test2Nodes500IterationTearDown(t *testing.T) {
	logFile, err := os.OpenFile("test_repeated_start_stop.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	require.NoError(t, err)
	defer logFile.Close()

	multiWriter := io.MultiWriter(os.Stdout, logFile)
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "ts"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderCfg),
		zapcore.AddSync(multiWriter),
		zap.DebugLevel,
	)
	customLogger := zap.New(core)
	SetLogger(customLogger)

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	initialMem := memStats.HeapAlloc
	Debug("[%s] Memory usage at test START: %d KB", t.Name(), initialMem/1024)

	initialRSS, err := utils.GetRSSKB()
	require.NoError(t, err)
	Debug("[%s] OS-level RSS at test START: %d KB", t.Name(), initialRSS)

	totalIterations := 500
	for i := 1; i <= totalIterations; i++ {
		var nodes []*WakuNode
		for n := 1; n <= 2; n++ {
			cfg := DefaultWakuConfig
			cfg.Relay = true
			cfg.Discv5Discovery = false
			cfg.TcpPort, cfg.Discv5UdpPort, err = GetFreePortIfNeeded(0, 0)
			require.NoError(t, err, "Failed to get free ports for node%d", n)
			node, err := NewWakuNode(&cfg, fmt.Sprintf("node%d", n))
			require.NoError(t, err, "Failed to create node%d", n)
			err = node.Start()
			require.NoError(t, err, "Failed to start node%d", n)
			nodes = append(nodes, node)
		}
		err = ConnectAllPeers(nodes)
		require.NoError(t, err)
		message := nodes[0].CreateMessage()
		msgHash, err := nodes[0].RelayPublishNoCTX(DefaultPubsubTopic, message)
		require.NoError(t, err)
		time.Sleep(500 * time.Millisecond)
		err = nodes[1].VerifyMessageReceived(message, msgHash, 500*time.Millisecond)
		require.NoError(t, err, "Node1 did not receive message from node1")
		for _, node := range nodes {
			node.StopAndDestroy()
		}
		runtime.GC()
		time.Sleep(250 * time.Millisecond)
		runtime.GC()
		if i == 250 || i == 500 {
			runtime.ReadMemStats(&memStats)
			Debug("Iteration %d, usage after teardown: %d KB", i, memStats.HeapAlloc/1024)
			require.LessOrEqual(t, memStats.HeapAlloc, initialMem*3, "Memory usage soared above threshold after iteration %d", i)
			rssNow, err := utils.GetRSSKB()
			require.NoError(t, err)
			Debug("Iteration %d, OS-level RSS after teardown: %d KB", i, rssNow)
			//require.LessOrEqual(t, rssNow, initialRSS*10, "OS-level RSS soared above threshold after iteration %d", i)
		}
		Debug("Iteration numberrrrrr  %d", i)
	}
	runtime.GC()
	time.Sleep(500 * time.Millisecond)
	runtime.GC()
	runtime.ReadMemStats(&memStats)
	finalMem := memStats.HeapAlloc
	Debug("[%s] Memory usage at test END: %d KB", t.Name(), finalMem/1024)
	require.LessOrEqual(t, finalMem, initialMem*3, "Memory usage soared above threshold after %d cycles", totalIterations)
	finalRSS, err := utils.GetRSSKB()
	require.NoError(t, err)
	Debug("[%s] OS-level RSS at test END: %d KB", t.Name(), finalRSS)
	require.LessOrEqual(t, finalRSS, initialRSS*3, "OS-level RSS soared above threshold after %d cycles", totalIterations)
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

	Debug("[%s] Test completed successfully", t.Name())
}

func TestHighThroughput10kPublish(t *testing.T) {

	node1Cfg := DefaultWakuConfig
	node1Cfg.Relay = true
	node1, err := StartWakuNode("node1", &node1Cfg)
	require.NoError(t, err, "Failed to start node1")
	defer node1.StopAndDestroy()

	node2Cfg := DefaultWakuConfig
	node2Cfg.Relay = true
	node2, err := StartWakuNode("node2", &node2Cfg)
	require.NoError(t, err, "Failed to start node2")
	defer node2.StopAndDestroy()

	err = node1.ConnectPeer(node2)
	require.NoError(t, err, "Failed to connect node1 to node2")

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	startHeapKB := memStats.HeapAlloc / 1024
	startRSSKB, err := utils.GetRSSKB()
	require.NoError(t, err, "Failed to read initial RSS")

	Debug("Memory usage BEFORE sending => HeapAlloc: %d KB, RSS: %d KB", startHeapKB, startRSSKB)

	totalMessages := 10000
	pubsubTopic := DefaultPubsubTopic

	startTime := time.Now()
	for i := 0; i < totalMessages; i++ {
		message := node1.CreateMessage()
		message.Payload = []byte(fmt.Sprintf("High-throughput message #%d", i))

		_, err := node1.RelayPublishNoCTX(pubsubTopic, message)
		require.NoError(t, err, "Failed to publish message %d", i)
	}
	duration := time.Since(startTime)

	runtime.ReadMemStats(&memStats)
	endHeapKB := memStats.HeapAlloc / 1024
	endRSSKB, err := utils.GetRSSKB()
	require.NoError(t, err, "Failed to read final RSS")

	Debug("Memory usage AFTER sending => HeapAlloc: %d KB, RSS: %d KB", endHeapKB, endRSSKB)

	Debug("Published %d messages in %s", totalMessages, duration)
	Debug("Total time per message ~ %v", duration/time.Duration(totalMessages))
}

func TestConnectDisconnect50Iteration(t *testing.T) {
	node0Cfg := DefaultWakuConfig
	node0Cfg.Relay = true
	node0, err := StartWakuNode("node0", &node0Cfg)
	require.NoError(t, err, "Failed to start node0")

	node1Cfg := DefaultWakuConfig
	node1Cfg.Relay = true
	node1, err := StartWakuNode("node1", &node1Cfg)
	require.NoError(t, err, "Failed to start node1")

	defer func() {
		node0.StopAndDestroy()
		node1.StopAndDestroy()
	}()

	iterations := 50
	for i := 1; i <= iterations; i++ {
		if i%2 == 1 {
			err := node0.ConnectPeer(node1)
			require.NoError(t, err, "Iteration %d: node0 failed to connect to node1", i)
			time.Sleep(1 * time.Second)
			count, err := node0.GetNumConnectedPeers()
			require.NoError(t, err, "Iteration %d: failed to get peers for node0", i)
			Debug("Iteration %d: node0 sees %d connected peers", i, count)
			if count == 1 {
				msg := node0.CreateMessage()
				msg.Payload = []byte(fmt.Sprintf("Iteration %d: message from node0", i))
				msgHash, err := node0.RelayPublishNoCTX(DefaultPubsubTopic, msg)
				require.NoError(t, err, "Iteration %d: node0 failed to publish message", i)
				Debug("Iteration %d: node0 published message with hash %s", i, msgHash.String())
			}
			err = node0.DisconnectPeer(node1)
			require.NoError(t, err, "Iteration %d: node0 failed to disconnect from node1", i)
			Debug("Iteration %d: node0 disconnected from node1", i)
		} else {
			err := node1.ConnectPeer(node0)
			require.NoError(t, err, "Iteration %d: node1 failed to connect to node0", i)
			time.Sleep(1 * time.Second)
			count, err := node1.GetNumConnectedPeers()
			require.NoError(t, err, "Iteration %d: failed to get peers for node1", i)
			Debug("Iteration %d: node1 sees %d connected peers", i, count)
			if count == 1 {
				msg := node1.CreateMessage()
				msg.Payload = []byte(fmt.Sprintf("Iteration %d: message from node1", i))
				msgHash, err := node1.RelayPublishNoCTX(DefaultPubsubTopic, msg)
				require.NoError(t, err, "Iteration %d: node1 failed to publish message", i)
				Debug("Iteration %d: node1 published message with hash %s", i, msgHash.String())
			}
			err = node1.DisconnectPeer(node0)
			require.NoError(t, err, "Iteration %d: node1 failed to disconnect from node0", i)
			Debug("Iteration %d: node1 disconnected from node0", i)
		}
		time.Sleep(2 * time.Second)
	}
}
