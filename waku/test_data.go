package waku

import (
	"sync"
	"time"
)

const ConnectPeerTimeout = 10 * time.Second //default timeout for node to connect to another node

var DefaultPubsubTopic = "/waku/2/rs/16/64"
var (
	MinPort    = 1024               // Minimum allowable port (exported)
	MaxPort    = 65535              // Maximum allowable port (exported)
	usedPorts  = make(map[int]bool) // Tracks used ports (internal to package)
	portsMutex sync.Mutex           // Ensures thread-safe access to usedPorts
)

// Default configuration values
var DefaultWakuConfig = WakuConfig{
	Relay:           false,
	LogLevel:        "DEBUG",
	Discv5Discovery: true,
	ClusterID:       16,
	Shards:          []uint16{64},
	PeerExchange:    false,
	Store:           false,
	Filter:          false,
	Lightpush:       false,
	Discv5UdpPort:   GenerateUniquePort(),
	TcpPort:         GenerateUniquePort(),
}
