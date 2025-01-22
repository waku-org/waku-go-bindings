package testlibs

import (
	"context"
	"errors"

	utilities "github.com/waku-org/waku-go-bindings/testlibs/utilities"
	"github.com/waku-org/waku-go-bindings/waku"
	"go.uber.org/zap"
)

type WakuNodeWrapper struct {
	*waku.WakuNode
}

// This function create waku node from config and start it
func Wrappers_StartWakuNode(customCfg *waku.WakuConfig, logger *zap.Logger) (*WakuNodeWrapper, error) {

	var nodeCfg waku.WakuConfig

	if customCfg == nil {
		nodeCfg = *utilities.DefaultWakuConfig
	} else {
		nodeCfg = *customCfg
	}

	nodeCfg.Discv5UdpPort = utilities.GenerateUniquePort()
	nodeCfg.TcpPort = utilities.GenerateUniquePort()

	utilities.Debug("Create node with default config")
	node, err := waku.NewWakuNode(&nodeCfg, logger)
	if err != nil {
		utilities.Error("Can't create node")
		return nil, err
	}

	utilities.Debug("Attempting to start WakuNode")
	wrapper := &WakuNodeWrapper{WakuNode: node}
	if wrapper.WakuNode == nil {
		err := errors.New("WakuNode instance is nil")
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

// Stops the WakuNode .
func (node *WakuNodeWrapper) Wrappers_Stop() error {
	if node == nil || node.WakuNode == nil {
		err := errors.New("WakuNode instance is nil")
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

// Destroys the WakuNode .
func (node *WakuNodeWrapper) Wrappers_Destroy() error {
	if node == nil || node.WakuNode == nil {
		err := errors.New("WakuNode instance is nil")
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

func (wrapper *WakuNodeWrapper) Wrappers_StopAndDestroy() error {
	if wrapper.WakuNode == nil {
		err := errors.New("WakuNode instance is nil")
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

func (wrapper *WakuNodeWrapper) Wrappers_GetNumConnectedRelayPeers(optPubsubTopic ...string) (int, error) {
	utilities.Debug("Wrappers_GetNumConnectedRelayPeers called")

	if wrapper.WakuNode == nil {
		err := errors.New("WakuNode is nil in WakuNodeWrapper")
		utilities.Error("Cannot proceed; node is nil", zap.Error(err))
		// Return an error immediately to “stop” the function
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

func (w *WakuNodeWrapper) Wrappers_ConnectPeer(target *WakuNodeWrapper) error {
	if w.WakuNode == nil {
		err := errors.New("WakuNode is nil in caller")
		utilities.Error("Cannot call Connect; caller node is nil", zap.Error(err))
		return err
	}
	if target == nil || target.WakuNode == nil {
		err := errors.New("target WakuNode is nil")
		utilities.Error("Cannot connect; target node is nil", zap.Error(err))
		return err
	}

	addrs, err := target.ListenAddresses()
	if err != nil || len(addrs) == 0 {
		errMsg := "failed to obtain target node's listening addresses"
		utilities.Error(errMsg, zap.String("error", err.Error()))
		return errors.New(errMsg)
	}

	peerAddr := addrs[0]

	utilities.Debug("Wrappers_ConnectPeer called", zap.String("targetAddr", peerAddr.String()))

	ctx, cancel := context.WithTimeout(context.Background(), utilities.ConnectPeerTimeout)
	defer cancel()

	utilities.Debug("Connecting to peer with address", zap.String("address", peerAddr.String()))
	err = w.WakuNode.Connect(ctx, peerAddr)
	if err != nil {
		utilities.Error("Failed to connect", zap.Error(err))
		return err
	}

	utilities.Debug("Successfully connected", zap.String("address", peerAddr.String()))
	return nil
}

func (wrapper *WakuNodeWrapper) Wrappers_DisconnectPeer(targetNode *WakuNodeWrapper) error {
	if wrapper.WakuNode == nil {
		err := errors.New("the calling WakuNode is nil")
		utilities.Error("Cannot disconnect; calling node is nil", zap.Error(err))
		return err
	}
	if targetNode == nil || targetNode.WakuNode == nil {
		err := errors.New("the target WakuNode is nil")
		utilities.Error("Cannot disconnect; target node is nil", zap.Error(err))
		return err
	}

	peerID, err := targetNode.WakuNode.PeerID()
	if err != nil {
		utilities.Error("Failed to retrieve peer ID from target node", zap.Error(err))
		return err
	}

	utilities.Debug("Wrappers_DisconnectPeer", zap.String("peerID", peerID.String()))
	err = wrapper.WakuNode.DisconnectPeerByID(peerID)
	if err != nil {
		utilities.Error("Failed to disconnect peer", zap.Error(err))
		return err
	}

	utilities.Debug("Successfully disconnected peer", zap.String("peerID", peerID.String()))
	return nil
}
