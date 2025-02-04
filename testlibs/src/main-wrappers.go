package testlibs

import (
	"context"
	"errors"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/waku-org/go-waku/waku/v2/protocol/pb"
	utilities "github.com/waku-org/waku-go-bindings/testlibs/utilities"
	"github.com/waku-org/waku-go-bindings/waku"
	"github.com/waku-org/waku-go-bindings/waku/common"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type WakuNodeWrapper struct {
	*waku.WakuNode
	config *waku.WakuConfig
}

// This function create waku node from config and start it
func StartWakuNode(customCfg *waku.WakuConfig, logger *zap.Logger) (*WakuNodeWrapper, error) {
	var nodeCfg waku.WakuConfig

	if customCfg == nil {
		nodeCfg = *utilities.DefaultWakuConfig
	} else {
		nodeCfg = *customCfg
	}

	nodeCfg.Discv5UdpPort = utilities.GenerateUniquePort()
	nodeCfg.TcpPort = utilities.GenerateUniquePort()

	utilities.Debug("Create node successfully")
	node, err := waku.NewWakuNode(&nodeCfg, logger)
	if err != nil {
		utilities.Error("Can't create node", zap.Error(err))
		return nil, err
	}

	utilities.Debug("Attempting to start WakuNode")
	wrapper := &WakuNodeWrapper{WakuNode: node, config: &nodeCfg} // Store config in wrapper

	if err := utilities.CheckWakuNodeNull(logger, wrapper.WakuNode); err != nil {
		utilities.Error("Failed to start WakuNode", zap.Error(err))
		return nil, err
	}

	err = wrapper.WakuNode.Start()
	if err != nil {
		utilities.Error("Failed to start WakuNode", zap.Error(err))
		return nil, err
	}
	utilities.Debug("Successfully started WakuNode")

	return wrapper, nil
}

// Stops the WakuNode.
func (node *WakuNodeWrapper) Stop() error {
	if err := utilities.CheckWakuNodeNull(nil, node.WakuNode); err != nil {
		utilities.Error("Failed to stop WakuNode", zap.Error(err))
		return err
	}

	utilities.Debug("Attempting to stop WakuNode")
	err := node.WakuNode.Stop()
	if err != nil {
		utilities.Error("Failed to stop WakuNode", zap.Error(err))
		return err
	}

	utilities.Debug("Successfully stopped WakuNode")
	return nil
}

// Destroys the WakuNode.
func (node *WakuNodeWrapper) Destroy() error {
	if err := utilities.CheckWakuNodeNull(nil, node.WakuNode); err != nil {
		utilities.Error("Failed to destroy WakuNode", zap.Error(err))
		return err
	}

	utilities.Debug("Attempting to destroy WakuNode")
	err := node.WakuNode.Destroy()
	if err != nil {
		utilities.Error("Failed to destroy WakuNode", zap.Error(err))
		return err
	}

	utilities.Debug("Successfully destroyed WakuNode")
	return nil
}

func (wrapper *WakuNodeWrapper) StopAndDestroy() error {
	if err := utilities.CheckWakuNodeNull(nil, wrapper.WakuNode); err != nil {
		utilities.Error("Failed to stop or destroy WakuNode", zap.Error(err))
		return err
	}

	utilities.Debug("Attempting to stop WakuNode")
	err := wrapper.Stop()
	if err != nil {
		utilities.Error("Failed to stop WakuNode", zap.Error(err))
		return err
	}

	utilities.Debug("Attempting to destroy WakuNode")
	err = wrapper.Destroy()
	if err != nil {
		utilities.Error("Failed to destroy WakuNode", zap.Error(err))
		return err
	}

	utilities.Debug("Successfully stopped and destroyed WakuNode")
	return nil
}

func (wrapper *WakuNodeWrapper) GetConnectedPeers() ([]peer.ID, error) {
	if err := utilities.CheckWakuNodeNull(nil, wrapper.WakuNode); err != nil {
		utilities.Error("Cannot proceed; node is nil", zap.Error(err))
		return nil, err
	}

	peerID, err := wrapper.WakuNode.PeerID()
	if err != nil {
		utilities.Error("Failed to get PeerID of node", zap.Error(err))
		return nil, err
	}

	utilities.Debug("Getting number of connected peers to node", zap.String("node", peerID.String()))

	peers, err := wrapper.WakuNode.GetConnectedPeers()
	if err != nil {
		utilities.Error("Failed to get connected peers", zap.Error(err))
		return nil, err
	}

	utilities.Debug("Successfully fetched connected peers",
		zap.Int("count", len(peers)),
	)
	return peers, nil
}

func (wrapper *WakuNodeWrapper) GetNumConnectedRelayPeers(optPubsubTopic ...string) (int, error) {
	utilities.Debug("GetNumConnectedRelayPeers called")

	if err := utilities.CheckWakuNodeNull(nil, wrapper.WakuNode); err != nil {
		utilities.Error("Cannot proceed; node is nil", zap.Error(err))
		return 0, err
	}

	numPeers, err := wrapper.WakuNode.GetNumConnectedRelayPeers(optPubsubTopic...)
	if err != nil {
		utilities.Error("Failed to get number of connected relay peers", zap.Error(err))
		return 0, err
	}

	utilities.Debug("Successfully fetched number of connected relay peers",
		zap.Int("count", numPeers),
	)
	return numPeers, nil
}

func (wrapper *WakuNodeWrapper) ConnectPeer(targetNode *WakuNodeWrapper) error {

	utilities.Debug("Connect node to peer")
	if err := utilities.CheckWakuNodeNull(nil, wrapper.WakuNode); err != nil {
		utilities.Error("Cannot call Connect; caller node is nil", zap.Error(err))
		return err
	}
	if err := utilities.CheckWakuNodeNull(nil, targetNode.WakuNode); err != nil {
		utilities.Error("Cannot connect; target node is nil", zap.Error(err))
		return err
	}

	targetPeerID, err := targetNode.WakuNode.PeerID()
	if err != nil {
		utilities.Error("Failed to get PeerID of target node", zap.Error(err))
		return err
	}

	utilities.Debug("Get connected peers before attempting to connect")

	connectedPeersBefore, err := wrapper.GetConnectedPeers()
	if err != nil {
		utilities.Debug("Could not fetch connected peers before connecting (might be none yet)", zap.Error(err))
	} else {
		utilities.Debug("Connected peers before connecting", zap.Int("count", len(connectedPeersBefore)))
	}

	utilities.Debug("Attempt to connect to the target node")
	ctx, cancel := context.WithTimeout(context.Background(), utilities.ConnectPeerTimeout)
	defer cancel()

	targetAddr, err := targetNode.WakuNode.ListenAddresses()
	if err != nil || len(targetAddr) == 0 {
		utilities.Error("Failed to get listen addresses for target node", zap.Error(err))
		return errors.New("target node has no listen addresses")
	}

	utilities.Debug("Connecting to peer", zap.String("address", targetAddr[0].String()))
	err = wrapper.WakuNode.Connect(ctx, targetAddr[0])
	if err != nil {
		utilities.Error("Failed to connect to peer", zap.Error(err))
		return err
	}

	utilities.Debug("Get connected peers after attempting to connect")
	connectedPeersAfter, err := wrapper.GetConnectedPeers()
	if err != nil {
		utilities.Error("Failed to get connected peers after connecting", zap.Error(err))
		return err
	}

	utilities.Debug("Connected peers after connecting", zap.Int("count", len(connectedPeersAfter)))

	utilities.Debug("Check if the target peer is now connected")
	isConnected := false
	for _, peerID := range connectedPeersAfter {
		if peerID == targetPeerID {
			isConnected = true
			break
		}
	}

	if !isConnected {
		err := errors.New("failed to connect; target peer is not in connected peers list")
		utilities.Error("Connect operation failed", zap.Error(err))
		return err
	}

	utilities.Debug("Successfully connected to target peer", zap.String("targetPeerID", targetPeerID.String()))
	return nil
}

func (wrapper *WakuNodeWrapper) DisconnectPeer(target *WakuNodeWrapper) error {

	if err := utilities.CheckWakuNodeNull(nil, wrapper.WakuNode); err != nil {
		utilities.Error("Cannot call Disconnect; caller node is nil", zap.Error(err))
		return err
	}
	if err := utilities.CheckWakuNodeNull(nil, target.WakuNode); err != nil {
		utilities.Error("Cannot disconnect; target node is nil", zap.Error(err))
		return err
	}

	utilities.Debug("Check if nodes are peers first")

	peerID, err := target.WakuNode.PeerID()
	if err != nil {
		utilities.Error("Failed to get PeerID of target node", zap.Error(err))
		return err
	}

	connectedPeers, err := wrapper.GetConnectedPeers()
	if err != nil {
		utilities.Error("Failed to get connected peers", zap.Error(err))
		return err
	}

	isPeer := false
	for _, connectedPeerID := range connectedPeers {
		if connectedPeerID == peerID {
			isPeer = true
			break
		}
	}

	if !isPeer {
		err = errors.New("nodes are not connected as peers")
		utilities.Error("Cannot disconnect; nodes are not peers", zap.Error(err))
		return err
	}

	utilities.Debug("Nodes are peers.. attempting to disconnect")
	err = wrapper.WakuNode.DisconnectPeerByID(peerID)
	if err != nil {
		utilities.Error("Failed to disconnect peer", zap.Error(err))
		return err
	}

	utilities.Debug("Successfully disconnected peer", zap.String("peerID", peerID.String()))
	return nil
}

// Wrapper for Peer Exchange Request
func (wrapper *WakuNodeWrapper) PeerExchangeRequest(numPeers uint64) (uint64, error) {
	if err := utilities.CheckWakuNodeNull(nil, wrapper.WakuNode); err != nil {
		utilities.Error("Cannot perform peer exchange; node is nil", zap.Error(err))
		return 0, err
	}

	utilities.Debug("Attempting peer exchange request", zap.Uint64("numPeers", numPeers))
	numRecvPeers, err := wrapper.WakuNode.PeerExchangeRequest(numPeers)
	if err != nil {
		utilities.Error("Failed to perform peer exchange request", zap.Error(err))
		return 0, err
	}

	utilities.Debug("Successfully received peers from peer exchange", zap.Uint64("numPeersReceived", numRecvPeers))
	return numRecvPeers, nil
}

// Wrapper for Start Discovery V5
func (wrapper *WakuNodeWrapper) StartDiscV5() error {
	if err := utilities.CheckWakuNodeNull(nil, wrapper.WakuNode); err != nil {
		utilities.Error("Cannot start Discovery V5; node is nil", zap.Error(err))
		return err
	}

	if wrapper.config == nil || !wrapper.config.Discv5Discovery {
		err := errors.New("discovery V5 is not enabled in the node configuration")
		utilities.Error("Failed to start Discovery V5", zap.Error(err))
		return err
	}
	utilities.Debug("Attempting to start Discovery V5")
	err := wrapper.WakuNode.StartDiscV5()
	if err != nil {
		utilities.Error("Failed to start Discovery V5", zap.Error(err))
		return err
	}

	utilities.Debug("Successfully started Discovery V5")
	return nil
}

// Wrapper for Stop Discovery V5
func (wrapper *WakuNodeWrapper) StopDiscV5() error {
	if err := utilities.CheckWakuNodeNull(nil, wrapper.WakuNode); err != nil {
		utilities.Error("Cannot stop Discovery V5; node is nil", zap.Error(err))
		return err
	}

	if wrapper.config == nil || !wrapper.config.Discv5Discovery {
		err := errors.New("discovery V5 is not enabled in the node configuration")
		utilities.Error("Failed to stop Discovery V5", zap.Error(err))
		return err
	}

	utilities.Debug("Attempting to stop Discovery V5")
	err := wrapper.WakuNode.StopDiscV5()
	if err != nil {
		utilities.Error("Failed to stop Discovery V5", zap.Error(err))
		return err
	}

	utilities.Debug("Successfully stopped Discovery V5")
	return nil
}

func (wrapper *WakuNodeWrapper) Version() (string, error) {
	if err := utilities.CheckWakuNodeNull(nil, wrapper.WakuNode); err != nil {
		utilities.Error("cannot get version; node is nil", zap.Error(err))
		return "", err
	}

	utilities.Debug("Attempting to retrieve Waku node version")
	version, err := wrapper.WakuNode.Version()
	if err != nil {
		utilities.Error("failed to retrieve Waku node version", zap.Error(err))
		return "", err
	}

	utilities.Debug("Successfully retrieved Waku node version", zap.String("version", version))
	return version, nil
}

func (wrapper *WakuNodeWrapper) CreateMessage(customMessage ...*pb.WakuMessage) *pb.WakuMessage {
	logger, _ := zap.NewDevelopment()
	logger.Debug("Creating a WakuMessage")

	if len(customMessage) > 0 && customMessage[0] != nil {
		logger.Debug("Using provided custom message")
		return customMessage[0]
	}

	logger.Debug("Using default message format")
	defaultMessage := &pb.WakuMessage{
		Payload:      []byte("This is a default Waku message payload"),
		ContentTopic: "test-content-topic",
		Version:      proto.Uint32(0),
		Timestamp:    proto.Int64(time.Now().UnixNano()),
	}

	logger.Debug("Successfully created a default WakuMessage")
	return defaultMessage
}

func (wrapper *WakuNodeWrapper) VerifyMessageReceived(expectedMessage *pb.WakuMessage, expectedHash common.MessageHash) error {
	logger, _ := zap.NewDevelopment()
	logger.Debug("Verifying if the message was received")

	select {
	case envelope := <-wrapper.MsgChan:
		if envelope == nil {
			logger.Debug("Received envelope is nil")
			return errors.New("received envelope is nil")
		}
		if string(expectedMessage.Payload) != string(envelope.Message().Payload) {
			logger.Debug("Payload does not match")
			return errors.New("payload does not match")
		}
		if expectedMessage.ContentTopic != envelope.Message().ContentTopic {
			logger.Debug("Content topic does not match")
			return errors.New("content topic does not match")
		}
		if expectedHash != envelope.Hash() {
			logger.Debug("Message hash does not match")
			return errors.New("message hash does not match")
		}
		logger.Debug("Message received and verified successfully")
		return nil
	case <-time.After(5 * time.Second):
		logger.Debug("Timeout: message not received within 5 seconds")
		return errors.New("timeout: message not received within 5 seconds")
	}
}

func ConnectAllPeers(nodes []*WakuNodeWrapper) error {
	logger, _ := zap.NewDevelopment()
	logger.Debug("Connecting nodes in a relay chain")

	for i := 0; i < len(nodes)-1; i++ {
		logger.Debug("Connecting node", zap.Int("from", i), zap.Int("to", i+1))
		err := nodes[i].ConnectPeer(nodes[i+1])
		if err != nil {
			logger.Debug("Failed to connect nodes", zap.Error(err))
			return err
		}
	}

	time.Sleep(4 * time.Second)
	logger.Debug("Waiting for connections to stabilize")

	return nil
}
