package waku_go_tests

import (
	"testing"

	testlibs "github.com/waku-org/waku-go-bindings/testlibs/src"
	"go.uber.org/zap"
)

func TestCreateTwoWakuNodes(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	node1, err := testlibs.Wrappers_CreateWakuNode(nil, logger)
	if err != nil {
		t.Fatalf("Failed to create node1: %v", err)
	}

	node2, err := testlibs.Wrappers_CreateWakuNode(nil, logger)
	if err != nil {
		t.Fatalf("Failed to create node2: %v", err)
	}

	if err := node1.Start(); err != nil {
		t.Fatalf("Failed to start node1: %v", err)
	}
	if err := node2.Start(); err != nil {
		t.Fatalf("Failed to start node2: %v", err)
	}

	node1.Stop()
	node2.Stop()
}
