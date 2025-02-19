package waku

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/cenkalti/backoff/v3"
	"github.com/waku-org/go-waku/waku/v2/protocol/pb"
	"github.com/waku-org/waku-go-bindings/waku/common"
	"google.golang.org/protobuf/proto"
)

type NwakuInfo struct {
	ListenAddresses []string `json:"listenAddresses"`
	EnrUri          string   `json:"enrUri"`
}

func GetNwakuInfo(host *string, port *int) (NwakuInfo, error) {
	nwakuRestPort := 8645
	if port != nil {
		nwakuRestPort = *port
	}
	envNwakuRestPort := os.Getenv("NWAKU_REST_PORT")
	if envNwakuRestPort != "" {
		v, err := strconv.Atoi(envNwakuRestPort)
		if err != nil {
			return NwakuInfo{}, err
		}
		nwakuRestPort = v
	}

	nwakuRestHost := "localhost"
	if host != nil {
		nwakuRestHost = *host
	}
	envNwakuRestHost := os.Getenv("NWAKU_REST_HOST")
	if envNwakuRestHost != "" {
		nwakuRestHost = envNwakuRestHost
	}

	resp, err := http.Get(fmt.Sprintf("http://%s:%d/debug/v1/info", nwakuRestHost, nwakuRestPort))
	if err != nil {
		return NwakuInfo{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return NwakuInfo{}, err
	}

	var data NwakuInfo
	err = json.Unmarshal(body, &data)
	if err != nil {
		return NwakuInfo{}, err
	}

	return data, nil
}

type BackOffOption func(*backoff.ExponentialBackOff)

func RetryWithBackOff(o func() error, options ...BackOffOption) error {
	b := backoff.ExponentialBackOff{
		InitialInterval:     time.Millisecond * 100,
		RandomizationFactor: 0.1,
		Multiplier:          1,
		MaxInterval:         time.Second,
		MaxElapsedTime:      time.Second * 10,
		Clock:               backoff.SystemClock,
	}
	for _, option := range options {
		option(&b)
	}
	b.Reset()
	return backoff.Retry(o, &b)
}

func (n *WakuNode) CreateMessage(customMessage ...*pb.WakuMessage) *pb.WakuMessage {
	Debug("Creating a WakuMessage on node %s", n.nodeName)

	if len(customMessage) > 0 && customMessage[0] != nil {
		Debug("Using provided custom message on node %s", n.nodeName)
		return customMessage[0]
	}

	Debug("Using default message format on node %s", n.nodeName)
	defaultMessage := &pb.WakuMessage{
		Payload:      []byte("This is a default Waku message payload"),
		ContentTopic: "test-content-topic",
		Version:      proto.Uint32(0),
		Timestamp:    proto.Int64(time.Now().UnixNano()),
	}

	Debug("Successfully created a default WakuMessage on node %s", n.nodeName)
	return defaultMessage
}

func WaitForAutoConnection(nodeList []*WakuNode) error {
	Debug("Waiting for auto-connection of nodes...")

	options := func(b *backoff.ExponentialBackOff) {
		b.MaxElapsedTime = 30 * time.Second
	}

	err := RetryWithBackOff(func() error {
		for _, node := range nodeList {
			peers, err := node.GetConnectedPeers()
			if err != nil {
				return err
			}

			if len(peers) < 1 {
				return errors.New("expected at least one connected peer") // Retry
			}

			Debug("Node %s has %d connected peers", node.nodeName, len(peers))
		}

		return nil
	}, options)

	if err != nil {
		Error("Auto-connection failed after retries: %v", err)
		return err
	}

	Debug("Auto-connection check completed successfully")
	return nil
}

func (n *WakuNode) VerifyMessageReceived(expectedMessage *pb.WakuMessage, expectedHash common.MessageHash, timeout ...time.Duration) error {

	var verifyTimeout time.Duration
	if len(timeout) > 0 {
		verifyTimeout = timeout[0]
	} else {
		verifyTimeout = DefaultTimeOut
	}

	Debug("Verifying if the message was received on node %s, timeout: %v", n.nodeName, verifyTimeout)

	ctx, cancel := context.WithTimeout(context.Background(), verifyTimeout)
	defer cancel()

	select {
	case envelope := <-n.MsgChan:
		if envelope == nil {
			Error("Received envelope is nil on node %s", n.nodeName)
			return errors.New("received envelope is nil")
		}
		if string(expectedMessage.Payload) != string(envelope.Message().Payload) {
			Error("Payload does not match on node %s", n.nodeName)
			return errors.New("payload does not match")
		}
		if expectedMessage.ContentTopic != envelope.Message().ContentTopic {
			Error("Content topic does not match on node %s", n.nodeName)
			return errors.New("content topic does not match")
		}
		if expectedHash != envelope.Hash() {
			Error("Message hash does not match on node %s", n.nodeName)
			return errors.New("message hash does not match")
		}
		Debug("Message received and verified successfully on node %s, Message: %s", n.nodeName, string(envelope.Message().Payload))
		return nil
	case <-ctx.Done():
		Error("Timeout: message not received within %v on node %s", verifyTimeout, n.nodeName)
		return errors.New("timeout: message not received within the given duration")
	}
}
