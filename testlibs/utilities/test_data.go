package utilities

import (
	"time"
)

var ConnectPeerTimeout = 10 * time.Second //default timeout for node to connect to another node

var DefaultPubsubTopic = "/waku/2/rs/3/0"
var PubsubTopic1 = "/waku/2/rs/3/1"
