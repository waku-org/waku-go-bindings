package utilities

import (
	"errors"
	"math/rand"
	"sync"
	"time"

	"github.com/waku-org/waku-go-bindings/waku"
	"go.uber.org/zap"
)

var (
	MinPort    = 1024               // Minimum allowable port (exported)
	MaxPort    = 65535              // Maximum allowable port (exported)
	usedPorts  = make(map[int]bool) // Tracks used ports (internal to package)
	portsMutex sync.Mutex           // Ensures thread-safe access to usedPorts
)

// Default configuration values
var DefaultWakuConfig = &waku.WakuConfig{
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

// WakuConfigOption is a function that applies a change to a WakuConfig.
type WakuConfigOption func(*waku.WakuConfig)

func GenerateUniquePort() int {
	rng := rand.New(rand.NewSource(time.Now().UnixNano())) // Local RNG instance

	for {
		port := rng.Intn(MaxPort-MinPort+1) + MinPort

		portsMutex.Lock()
		if !usedPorts[port] {
			usedPorts[port] = true
			portsMutex.Unlock()
			return port
		}
		portsMutex.Unlock()
	}
}

func CheckWakuNodeNull(logger *zap.Logger, node interface{}) error {
	if node == nil {
		err := errors.New("WakuNode instance is nil")
		if logger != nil {
			logger.Error("WakuNode is nil", zap.Error(err))
		}
		return err
	}
	return nil
}
