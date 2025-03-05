package waku

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/waku-org/waku-go-bindings/waku/common"
	"google.golang.org/protobuf/proto"
)

func TestMemoryUsageForThreeNodes(t *testing.T) {
	logFile, err := os.OpenFile("test_logs.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	require.NoError(t, err, "Failed to open log file")

	multiWriter := log.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)

	Debug("Starting memory usage test for three Waku nodes")

	var memStats runtime.MemStats

	runtime.ReadMemStats(&memStats)
	initialHeapAlloc := memStats.HeapAlloc
	Debug("Initial memory usage before creating nodes: %d MB", initialHeapAlloc/1024/1024)

	node1, err := NewWakuNode(&DefaultWakuConfig, "node1")
	require.NoError(t, err, "Failed to create Node1")

	node2, err := NewWakuNode(&DefaultWakuConfig, "node2")
	require.NoError(t, err, "Failed to create Node2")

	node3, err := NewWakuNode(&DefaultWakuConfig, "node3")
	require.NoError(t, err, "Failed to create Node3")

	runtime.ReadMemStats(&memStats)
	afterCreationHeapAlloc := memStats.HeapAlloc
	Debug("Memory usage after creating nodes: %d MB", afterCreationHeapAlloc/1024/1024)

	err = node1.Start()
	require.NoError(t, err, "Failed to start Node1")

	err = node2.Start()
	require.NoError(t, err, "Failed to start Node2")

	err = node3.Start()
	require.NoError(t, err, "Failed to start Node3")

	runtime.ReadMemStats(&memStats)
	afterStartHeapAlloc := memStats.HeapAlloc
	Debug("Memory usage after starting nodes: %d MB", afterStartHeapAlloc/1024/1024)

	time.Sleep(2 * time.Second)

	node1.Stop()
	node2.Stop()
	node3.Stop()

	runtime.ReadMemStats(&memStats)
	afterStopHeapAlloc := memStats.HeapAlloc
	Debug("Memory usage after stopping nodes: %d MB", afterStopHeapAlloc/1024/1024)

	require.LessOrEqual(t, afterStartHeapAlloc, initialHeapAlloc*3, "Unexpected memory growth after starting nodes")

	Debug("Test completed successfully")

	logFile.Close()
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
