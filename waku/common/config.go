package common

import (
	"encoding/json"
	"fmt"
)

type WakuConfig struct {
	Host                        string           `json:"host,omitempty"`
	Nodekey                     string           `json:"nodekey,omitempty"`
	Relay                       bool             `json:"relay"`
	Store                       bool             `json:"store,omitempty"`
	LegacyStore                 bool             `json:"legacyStore"`
	Storenode                   string           `json:"storenode,omitempty"`
	StoreMessageRetentionPolicy string           `json:"storeMessageRetentionPolicy,omitempty"`
	StoreMessageDbUrl           string           `json:"storeMessageDbUrl,omitempty"`
	StoreMessageDbVacuum        bool             `json:"storeMessageDbVacuum,omitempty"`
	StoreMaxNumDbConnections    int              `json:"storeMaxNumDbConnections,omitempty"`
	StoreResume                 bool             `json:"storeResume,omitempty"`
	Filter                      bool             `json:"filter,omitempty"`
	Filternode                  string           `json:"filternode,omitempty"`
	FilterSubscriptionTimeout   int64            `json:"filterSubscriptionTimeout,omitempty"`
	FilterMaxPeersToServe       uint32           `json:"filterMaxPeersToServe,omitempty"`
	FilterMaxCriteria           uint32           `json:"filterMaxCriteria,omitempty"`
	Lightpush                   bool             `json:"lightpush,omitempty"`
	LightpushNode               string           `json:"lightpushnode,omitempty"`
	LogLevel                    string           `json:"logLevel,omitempty"`
	DnsDiscovery                bool             `json:"dnsDiscovery,omitempty"`
	DnsDiscoveryUrl             string           `json:"dnsDiscoveryUrl,omitempty"`
	MaxMessageSize              string           `json:"maxMessageSize,omitempty"`
	Staticnodes                 []string         `json:"staticnodes,omitempty"`
	Discv5BootstrapNodes        []string         `json:"discv5BootstrapNodes,omitempty"`
	Discv5Discovery             bool             `json:"discv5Discovery,omitempty"`
	Discv5UdpPort               int              `json:"discv5UdpPort,omitempty"`
	ClusterID                   uint16           `json:"clusterId,omitempty"`
	Shards                      []uint16         `json:"shards,omitempty"`
	PeerExchange                bool             `json:"peerExchange,omitempty"`
	PeerExchangeNode            string           `json:"peerExchangeNode,omitempty"`
	TcpPort                     int              `json:"tcpPort,omitempty"`
	RateLimits                  RateLimitsConfig `json:"rateLimits,omitempty"`
}

type RateLimitsConfig struct {
	Filter       *RateLimit `json:"-"`
	Lightpush    *RateLimit `json:"-"`
	PeerExchange *RateLimit `json:"-"`
}

type RateLimit struct {
	Volume   int               // Number of allowed messages per period
	Period   int               // Length of each rate-limit period (in TimeUnit)
	TimeUnit RateLimitTimeUnit // Time unit of the period
}

type RateLimitTimeUnit string

const Hour RateLimitTimeUnit = "h"
const Minute RateLimitTimeUnit = "m"
const Second RateLimitTimeUnit = "s"
const Millisecond RateLimitTimeUnit = "ms"

func (rl RateLimit) String() string {
	return fmt.Sprintf("%d/%d%s", rl.Volume, rl.Period, rl.TimeUnit)
}

func (rl RateLimit) MarshalJSON() ([]byte, error) {
	return json.Marshal(rl.String())
}

func (rlc RateLimitsConfig) MarshalJSON() ([]byte, error) {
	output := []string{}
	if rlc.Filter != nil {
		output = append(output, fmt.Sprintf("filter:%s", rlc.Filter.String()))
	}
	if rlc.Lightpush != nil {
		output = append(output, fmt.Sprintf("lightpush:%s", rlc.Lightpush.String()))
	}
	if rlc.PeerExchange != nil {
		output = append(output, fmt.Sprintf("px:%s", rlc.PeerExchange.String()))
	}
	return json.Marshal(output)
}
