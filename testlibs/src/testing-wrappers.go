package testlibs

import (
	"errors"
     utilities "github.com/waku-org/waku-go-bindings/testlibs/utilities"
	"github.com/waku-org/waku-go-bindings/waku"
	"go.uber.org/zap"
)

type WakuNodeWrapper struct {
	*waku.WakuNode
}

func Wrappers_CreateWakuNode(customCfg *waku.WakuConfig, logger *zap.Logger) (*waku.WakuNode, error) {

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

	return node, nil
}

func (node *WakuNodeWrapper) Wrappers_Start() error {
	if node== nil || node.WakuNode == nil {
		err := errors.New("WakuNode instance is nil")
		utilities.Error("Failed to start WakuNode", zap.Error(err))
		return err
	}

	utilities.Debug("Attempting to start WakuNode")
	err := node.WakuNode.Start()
	if err != nil {
		utilities.Error("Failed to start WakuNode", zap.Error(err))
		return err
	}

	utilities.Debug("Successfully started WakuNode")
	return nil
}

// Stops the WakuNode instance.
func (node*WakuNodeWrapper) Wrappers_Stop() error {
	if node== nil || node.WakuNode == nil {
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

//  Destroys the WakuNode instance.
func (node*WakuNodeWrapper) Wrappers_Destroy() error {
	if node== nil || node.WakuNode == nil {
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
