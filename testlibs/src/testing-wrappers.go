package testlibs

import (
	"github.com/waku-org/waku-go-bindings/waku"
     
	"go.uber.org/zap"
)

func Wrappers_CreateWakuNode(customConfig *waku.WakuConfig, logger *zap.Logger) (*waku.WakuNode, error) {

	config := *DefaultWakuConfig
	config.Discv5UdpPort = GenerateUniquePort()
	config.TcpPort = GenerateUniquePort()

	if customConfig != nil {
		if customConfig.Relay {
			config.Relay = customConfig.Relay
		}
		if customConfig.LogLevel != "" {
			config.LogLevel = customConfig.LogLevel
		}
		if customConfig.Discv5Discovery {
			config.Discv5Discovery = customConfig.Discv5Discovery
		}
		if customConfig.ClusterID != 0 {
			config.ClusterID = customConfig.ClusterID
		}
		if len(customConfig.Shards) > 0 {
			config.Shards = customConfig.Shards
		}
		if customConfig.PeerExchange {
			config.PeerExchange = customConfig.PeerExchange
		}
	}

	node, err := waku.NewWakuNode(&config, logger)
	if err != nil {
		return nil, err
	}

	return node, nil
}
