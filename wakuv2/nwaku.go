package wakuv2

/*
	#cgo LDFLAGS: -L../third_party/nwaku/build/ -lnegentropy -lwaku
	#cgo LDFLAGS: -L../third_party/nwaku -Wl,-rpath,../third_party/nwaku/build/

	#include "../third_party/nwaku/library/libwaku.h"
	#include <stdio.h>
	#include <stdlib.h>

	extern void globalEventCallback(int ret, char* msg, size_t len, void* userData);

	typedef struct {
		int ret;
		char* msg;
		size_t len;
	} Resp;

	static void* allocResp() {
		return calloc(1, sizeof(Resp));
	}

	static void freeResp(void* resp) {
		if (resp != NULL) {
			free(resp);
		}
	}

	static char* getMyCharPtr(void* resp) {
		if (resp == NULL) {
			return NULL;
		}
		Resp* m = (Resp*) resp;
		return m->msg;
	}

	static size_t getMyCharLen(void* resp) {
		if (resp == NULL) {
			return 0;
		}
		Resp* m = (Resp*) resp;
		return m->len;
	}

	static int getRet(void* resp) {
		if (resp == NULL) {
			return 0;
		}
		Resp* m = (Resp*) resp;
		return m->ret;
	}

	// resp must be set != NULL in case interest on retrieving data from the callback
	static void callback(int ret, char* msg, size_t len, void* resp) {
		if (resp != NULL) {
			Resp* m = (Resp*) resp;
			m->ret = ret;
			m->msg = msg;
			m->len = len;
		}
	}

	#define WAKU_CALL(call)                                                        \
	do {                                                                           \
		int ret = call;                                                              \
		if (ret != 0) {                                                              \
			printf("Failed the call to: %s. Returned code: %d\n", #call, ret);         \
			exit(1);                                                                   \
		}                                                                            \
	} while (0)

	static void* cGoWakuNew(const char* configJson, void* resp) {
		// We pass NULL because we are not interested in retrieving data from this callback
		void* ret = waku_new(configJson, (WakuCallBack) callback, resp);
		return ret;
	}

	static void cGoWakuStart(void* wakuCtx, void* resp) {
		WAKU_CALL(waku_start(wakuCtx, (WakuCallBack) callback, resp));
	}

	static void cGoWakuStop(void* wakuCtx, void* resp) {
		WAKU_CALL(waku_stop(wakuCtx, (WakuCallBack) callback, resp));
	}

	static void cGoWakuDestroy(void* wakuCtx, void* resp) {
		WAKU_CALL(waku_destroy(wakuCtx, (WakuCallBack) callback, resp));
	}

	static void cGoWakuStartDiscV5(void* wakuCtx, void* resp) {
		WAKU_CALL(waku_start_discv5(wakuCtx, (WakuCallBack) callback, resp));
	}

	static void cGoWakuStopDiscV5(void* wakuCtx, void* resp) {
		WAKU_CALL(waku_stop_discv5(wakuCtx, (WakuCallBack) callback, resp));
	}

	static void cGoWakuVersion(void* wakuCtx, void* resp) {
		WAKU_CALL(waku_version(wakuCtx, (WakuCallBack) callback, resp));
	}

	static void cGoWakuSetEventCallback(void* wakuCtx) {
		// The 'globalEventCallback' Go function is shared amongst all possible Waku instances.

		// Given that the 'globalEventCallback' is shared, we pass again the
		// wakuCtx instance but in this case is needed to pick up the correct method
		// that will handle the event.

		// In other words, for every call the libwaku makes to globalEventCallback,
		// the 'userData' parameter will bring the context of the node that registered
		// that globalEventCallback.

		// This technique is needed because cgo only allows to export Go functions and not methods.

		waku_set_event_callback(wakuCtx, (WakuCallBack) globalEventCallback, wakuCtx);
	}

	static void cGoWakuContentTopic(void* wakuCtx,
							char* appName,
							int appVersion,
							char* contentTopicName,
							char* encoding,
							void* resp) {

		WAKU_CALL( waku_content_topic(wakuCtx,
							appName,
							appVersion,
							contentTopicName,
							encoding,
							(WakuCallBack) callback,
							resp) );
	}

	static void cGoWakuPubsubTopic(void* wakuCtx, char* topicName, void* resp) {
		WAKU_CALL( waku_pubsub_topic(wakuCtx, topicName, (WakuCallBack) callback, resp) );
	}

	static void cGoWakuDefaultPubsubTopic(void* wakuCtx, void* resp) {
		WAKU_CALL (waku_default_pubsub_topic(wakuCtx, (WakuCallBack) callback, resp));
	}

	static void cGoWakuRelayPublish(void* wakuCtx,
                       const char* pubSubTopic,
                       const char* jsonWakuMessage,
                       int timeoutMs,
					   void* resp) {

		WAKU_CALL (waku_relay_publish(wakuCtx,
                       pubSubTopic,
                       jsonWakuMessage,
                       timeoutMs,
                       (WakuCallBack) callback,
                       resp));
	}

	static void cGoWakuRelaySubscribe(void* wakuCtx, char* pubSubTopic, void* resp) {
		WAKU_CALL ( waku_relay_subscribe(wakuCtx,
							pubSubTopic,
							(WakuCallBack) callback,
							resp) );
	}

	static void cGoWakuRelayUnsubscribe(void* wakuCtx, char* pubSubTopic, void* resp) {

		WAKU_CALL ( waku_relay_unsubscribe(wakuCtx,
							pubSubTopic,
							(WakuCallBack) callback,
							resp) );
	}

	static void cGoWakuConnect(void* wakuCtx, char* peerMultiAddr, int timeoutMs, void* resp) {
		WAKU_CALL( waku_connect(wakuCtx,
						peerMultiAddr,
						timeoutMs,
						(WakuCallBack) callback,
						resp) );
	}

	static void cGoWakuDialPeer(void* wakuCtx,
									char* peerMultiAddr,
									char* protocol,
									int timeoutMs,
									void* resp) {

		WAKU_CALL( waku_dial_peer(wakuCtx,
						peerMultiAddr,
						protocol,
						timeoutMs,
						(WakuCallBack) callback,
						resp) );
	}

	static void cGoWakuDialPeerById(void* wakuCtx,
									char* peerId,
									char* protocol,
									int timeoutMs,
									void* resp) {

		WAKU_CALL( waku_dial_peer_by_id(wakuCtx,
						peerId,
						protocol,
						timeoutMs,
						(WakuCallBack) callback,
						resp) );
	}

	static void cGoWakuDisconnectPeerById(void* wakuCtx, char* peerId, void* resp) {
		WAKU_CALL( waku_disconnect_peer_by_id(wakuCtx,
						peerId,
						(WakuCallBack) callback,
						resp) );
	}

	static void cGoWakuListenAddresses(void* wakuCtx, void* resp) {
		WAKU_CALL (waku_listen_addresses(wakuCtx, (WakuCallBack) callback, resp) );
	}

	static void cGoWakuGetMyENR(void* ctx, void* resp) {
		WAKU_CALL (waku_get_my_enr(ctx, (WakuCallBack) callback, resp) );
	}

	static void cGoWakuGetMyPeerId(void* ctx, void* resp) {
		WAKU_CALL (waku_get_my_peerid(ctx, (WakuCallBack) callback, resp) );
	}

	static void cGoWakuPingPeer(void* ctx, char* peerAddr, int timeoutMs, void* resp) {
		WAKU_CALL (waku_ping_peer(ctx, peerAddr, timeoutMs, (WakuCallBack) callback, resp) );
	}

	static void cGoWakuListPeersInMesh(void* ctx, char* pubSubTopic, void* resp) {
		WAKU_CALL (waku_relay_get_num_peers_in_mesh(ctx, pubSubTopic, (WakuCallBack) callback, resp) );
	}

	static void cGoWakuGetNumConnectedRelayPeers(void* ctx, char* pubSubTopic, void* resp) {
		WAKU_CALL (waku_relay_get_num_connected_peers(ctx, pubSubTopic, (WakuCallBack) callback, resp) );
	}

	static void cGoWakuGetConnectedPeers(void* wakuCtx, void* resp) {
		WAKU_CALL (waku_get_connected_peers(wakuCtx, (WakuCallBack) callback, resp) );
	}

	static void cGoWakuGetPeerIdsFromPeerStore(void* wakuCtx, void* resp) {
		WAKU_CALL (waku_get_peerids_from_peerstore(wakuCtx, (WakuCallBack) callback, resp) );
	}

	static void cGoWakuLightpushPublish(void* wakuCtx,
					const char* pubSubTopic,
					const char* jsonWakuMessage,
					void* resp) {

		WAKU_CALL (waku_lightpush_publish(wakuCtx,
						pubSubTopic,
						jsonWakuMessage,
						(WakuCallBack) callback,
						resp));
	}

	static void cGoWakuStoreQuery(void* wakuCtx,
					const char* jsonQuery,
					const char* peerAddr,
					int timeoutMs,
					void* resp) {

		WAKU_CALL (waku_store_query(wakuCtx,
									jsonQuery,
									peerAddr,
									timeoutMs,
									(WakuCallBack) callback,
									resp));
	}

	static void cGoWakuPeerExchangeQuery(void* wakuCtx,
								uint64_t numPeers,
								void* resp) {

		WAKU_CALL (waku_peer_exchange_request(wakuCtx,
									numPeers,
									(WakuCallBack) callback,
									resp));
	}

	static void cGoWakuGetPeerIdsByProtocol(void* wakuCtx,
									 const char* protocol,
									 void* resp) {

		WAKU_CALL (waku_get_peerids_by_protocol(wakuCtx,
									protocol,
									(WakuCallBack) callback,
									resp));
	}

	static void cGoWakuDnsDiscovery(void* wakuCtx,
									 const char* entTreeUrl,
									 const char* nameDnsServer,
									 int timeoutMs,
									 void* resp) {

		WAKU_CALL (waku_dns_discovery(wakuCtx,
									entTreeUrl,
									nameDnsServer,
									timeoutMs,
									(WakuCallBack) callback,
									resp));
	}

*/
import "C"
import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/libp2p/go-libp2p/core/peer"
	libp2pproto "github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
	"github.com/waku-org/go-waku/waku/v2/protocol/pb"
	storepb "github.com/waku-org/go-waku/waku/v2/protocol/store/pb"
	"github.com/waku-org/go-waku/waku/v2/utils"
	"go.uber.org/zap"
)

const requestTimeout = 30 * time.Second

type WakuConfig struct {
	Host                 string   `json:"host,omitempty"`
	Port                 int      `json:"port,omitempty"`
	NodeKey              string   `json:"key,omitempty"`
	EnableRelay          bool     `json:"relay"`
	LogLevel             string   `json:"logLevel"`
	DnsDiscovery         bool     `json:"dnsDiscovery,omitempty"`
	DnsDiscoveryUrl      string   `json:"dnsDiscoveryUrl,omitempty"`
	MaxMessageSize       string   `json:"maxMessageSize,omitempty"`
	Staticnodes          []string `json:"staticnodes,omitempty"`
	Discv5BootstrapNodes []string `json:"discv5BootstrapNodes,omitempty"`
	Discv5Discovery      bool     `json:"discv5Discovery,omitempty"`
	Discv5UdpPort        uint16   `json:"discv5UdpPort,omitempty"`
	ClusterID            uint16   `json:"clusterId,omitempty"`
	Shards               []uint16 `json:"shards,omitempty"`
	PeerExchange         bool     `json:"peerExchange,omitempty"`
	PeerExchangeNode     string   `json:"peerExchangeNode,omitempty"`
	TcpPort              uint16   `json:"tcpPort,omitempty"`
}

// Waku represents a dark communication interface through the Ethereum
// network, using its very own P2P communication layer.
type Waku struct {
	node *WakuNode

	ctx    context.Context
	cancel context.CancelFunc

	wakuCfg *WakuConfig

	logger *zap.Logger
}

// Start implements node.Service, starting the background data propagation thread
// of the Waku protocol.
func (w *Waku) Start() error {
	err := w.node.Start()
	if err != nil {
		return fmt.Errorf("failed to start nwaku node: %v", err)
	}

	peerID, err := w.node.PeerID()
	if err != nil {
		return err
	}

	w.logger.Info("WakuV2 PeerID", zap.Stringer("id", peerID))

	return nil
}

// Stop implements node.Service, stopping the background data propagation thread
// of the Waku protocol.
func (w *Waku) Stop() error {
	w.cancel()

	err := w.node.Stop()
	if err != nil {
		return err
	}

	// w.wg.Wait()

	w.ctx = nil
	w.cancel = nil

	return nil
}

func (w *Waku) PeerCount() (int, error) {
	return w.node.GetNumConnectedPeers()
}

func (w *Waku) DialPeer(address multiaddr.Multiaddr) error {
	// Using WakuConnect so it matches the go-waku's behavior and terminology
	ctx, cancel := context.WithTimeout(w.ctx, requestTimeout)
	defer cancel()
	return w.node.Connect(ctx, address)
}

func (w *Waku) DialPeerByID(peerID peer.ID, protocol libp2pproto.ID) error {
	ctx, cancel := context.WithTimeout(w.ctx, requestTimeout)
	defer cancel()
	return w.node.DialPeerByID(ctx, peerID, protocol)
}

func (w *Waku) DropPeer(peerID peer.ID) error {
	return w.node.DisconnectPeerByID(peerID)
}

type request struct {
	id         string
	reqType    requestType
	input      any
	responseCh chan response
}

type response struct {
	err   error
	value any
}

// WakuNode represents an instance of an nwaku node
type WakuNode struct {
	wakuCtx   unsafe.Pointer
	cancel    context.CancelFunc
	requestCh chan *request
}

type requestType int

const (
	requestTypeNew requestType = iota + 1
	requestTypePing
	requestTypeStart
	requestTypeRelayPublish
	requestTypeStoreQuery
	requestTypePeerID
	requestTypeStop
	requestTypeDestroy
	requestTypeStartDiscV5
	requestTypeStopDiscV5
	requestTypeVersion
	requestTypeRelaySubscribe
	requestTypeRelayUnsubscribe
	requestTypePeerExchangeRequest
	requestTypeConnect
	requestTypeDialPeerByID
	requestTypeListenAddresses
	requestTypeENR
	requestTypeListPeersInMesh
	requestTypeGetConnectedPeers
	requestTypeGetPeerIDsFromPeerStore
	requestTypeGetPeerIDsByProtocol
	requestTypeDisconnectPeerByID
	requestTypeDnsDiscovery
	requestTypeDialPeer
	requestTypeGetNumConnectedRelayPeers
)

func newWakuNode(ctx context.Context, config *WakuConfig) (*WakuNode, error) {
	ctx, cancel := context.WithCancel(ctx)

	n := &WakuNode{
		requestCh: make(chan *request),
		cancel:    cancel,
	}

	// Notice this runs insto a separate goroutine. This is because we can't be sure
	// from which OS thread will go call nwaku operations (They need to be done from
	// the same thread that started nwaku). Communication with the goroutine to send
	// operations to nwaku will be done via channels
	go func() {
		// defer gocommon.LogOnPanic() TODO-nwaku

		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		C.waku_setup()

		n.processLoop(ctx)
	}()

	_, err := n.postTask(requestTypeNew, config)
	if err != nil {
		cancel()
		return nil, err
	}
	return n, nil
}

// New creates a WakuV2 client ready to communicate through the LibP2P network.
func New(nwakuCfg *WakuConfig, logger *zap.Logger) (*Waku, error) {
	var err error
	if logger == nil {
		logger, err = zap.NewDevelopment()
		if err != nil {
			return nil, err
		}
	}

	logger.Info("starting wakuv2 with config", zap.Any("nwakuCfg", nwakuCfg))

	ctx, cancel := context.WithCancel(context.Background())
	wakunode, err := newWakuNode(ctx, nwakuCfg)
	if err != nil {
		cancel()
		return nil, err
	}

	return &Waku{
		node:    wakunode,
		wakuCfg: nwakuCfg,
		logger:  logger,
		ctx:     ctx,
		cancel:  cancel,
	}, nil
}

func (n *WakuNode) postTask(reqType requestType, input any) (any, error) {
	responseCh := make(chan response)
	n.requestCh <- &request{
		reqType:    reqType,
		input:      input,
		responseCh: responseCh,
	}
	response := <-responseCh
	if response.err != nil {
		return nil, response.err
	}
	return response.value, nil
}

//export globalEventCallback
func globalEventCallback(callerRet C.int, msg *C.char, len C.size_t, userData unsafe.Pointer) {
	// This is shared among all Golang instances
	// TODO-nwaku
	// self := Waku{wakuCtx: userData}
	// self.MyEventCallback(callerRet, msg, len)
}

func (n *WakuNode) processLoop(ctx context.Context) {
	for req := range n.requestCh {
		switch req.reqType {
		case requestTypeNew:
			req.responseCh <- response{err: n.newNode(req.input.(*WakuConfig))}
		case requestTypePing:
			duration, err := n.pingPeer(req.input.(pingRequest))
			req.responseCh <- response{value: duration, err: err}
		case requestTypeStart:
			req.responseCh <- response{err: n.start()}
		case requestTypeRelayPublish:
			hash, err := n.relayPublish(req.input.(relayPublishRequest))
			req.responseCh <- response{value: hash, err: err}
		case requestTypeStoreQuery:
			results, err := n.storeQuery(req.input.(storeQueryRequest))
			req.responseCh <- response{value: results, err: err}
		case requestTypeDestroy:
			req.responseCh <- response{err: n.destroy()}
		case requestTypePeerID:
			peerID, err := n.peerID()
			req.responseCh <- response{value: peerID, err: err}
		case requestTypeStop:
			req.responseCh <- response{err: n.stop()}
		case requestTypeStartDiscV5:
			req.responseCh <- response{err: n.startDiscV5()}
		case requestTypeStopDiscV5:
			req.responseCh <- response{err: n.stopDiscV5()}
		case requestTypeVersion:
			version, err := n.version()
			req.responseCh <- response{value: version, err: err}
		case requestTypePeerExchangeRequest:
			numPeers, err := n.peerExchangeRequest(req.input.(uint64))
			req.responseCh <- response{value: numPeers, err: err}
		case requestTypeRelaySubscribe:
			req.responseCh <- response{err: n.relaySubscribe(req.input.(string))}
		case requestTypeRelayUnsubscribe:
			req.responseCh <- response{err: n.relayUnsubscribe(req.input.(string))}
		case requestTypeConnect:
			req.responseCh <- response{err: n.connect(req.input.(connectRequest))}
		case requestTypeDialPeerByID:
			req.responseCh <- response{err: n.dialPeerById(req.input.(dialPeerByIDRequest))}
		case requestTypeListenAddresses:
			addrs, err := n.listenAddresses()
			req.responseCh <- response{value: addrs, err: err}
		case requestTypeENR:
			enr, err := n.enr()
			req.responseCh <- response{value: enr, err: err}
		case requestTypeListPeersInMesh:
			numPeers, err := n.listPeersInMesh(req.input.(string))
			req.responseCh <- response{value: numPeers, err: err}
		case requestTypeGetConnectedPeers:
			peers, err := n.getConnectedPeers()
			req.responseCh <- response{value: peers, err: err}
		case requestTypeGetPeerIDsFromPeerStore:
			peers, err := n.getPeerIDsFromPeerStore()
			req.responseCh <- response{value: peers, err: err}
		case requestTypeGetPeerIDsByProtocol:
			peers, err := n.getPeerIDsByProtocol(req.input.(libp2pproto.ID))
			req.responseCh <- response{value: peers, err: err}
		case requestTypeDisconnectPeerByID:
			req.responseCh <- response{err: n.disconnectPeerByID(req.input.(peer.ID))}
		case requestTypeDnsDiscovery:
			addrs, err := n.dnsDiscovery(req.input.(dnsDiscoveryRequest))
			req.responseCh <- response{value: addrs, err: err}
		case requestTypeDialPeer:
			req.responseCh <- response{err: n.dialPeer(req.input.(dialPeerRequest))}
		case requestTypeGetNumConnectedRelayPeers:
			numPeers, err := n.getNumConnectedRelayPeers(req.input.([]string)...)
			req.responseCh <- response{value: numPeers, err: err}
		default:
			req.responseCh <- response{err: errors.New("invalid operation")}
		}
	}
}

func (n *WakuNode) getNumConnectedRelayPeers(optPubsubTopic ...string) (int, error) {
	var pubsubTopic string
	if len(optPubsubTopic) == 0 {
		pubsubTopic = ""
	} else {
		pubsubTopic = optPubsubTopic[0]
	}

	var resp = C.allocResp()
	var cPubsubTopic = C.CString(pubsubTopic)
	defer C.freeResp(resp)
	defer C.free(unsafe.Pointer(cPubsubTopic))

	C.cGoWakuGetNumConnectedRelayPeers(n.wakuCtx, cPubsubTopic, resp)

	if C.getRet(resp) == C.RET_OK {
		numPeersStr := C.GoStringN(C.getMyCharPtr(resp), C.int(C.getMyCharLen(resp)))
		numPeers, err := strconv.Atoi(numPeersStr)
		if err != nil {
			errMsg := "GetNumConnectedRelayPeers - error converting string to int: " + err.Error()
			return 0, errors.New(errMsg)
		}
		return numPeers, nil
	}
	errMsg := "error GetNumConnectedRelayPeers: " +
		C.GoStringN(C.getMyCharPtr(resp), C.int(C.getMyCharLen(resp)))
	return 0, errors.New(errMsg)
}

func (n *WakuNode) disconnectPeerByID(peerID peer.ID) error {
	var resp = C.allocResp()
	var cPeerId = C.CString(peerID.String())
	defer C.freeResp(resp)
	defer C.free(unsafe.Pointer(cPeerId))

	C.cGoWakuDisconnectPeerById(n.wakuCtx, cPeerId, resp)

	if C.getRet(resp) == C.RET_OK {
		return nil
	}
	errMsg := "error DisconnectPeerById: " + C.GoStringN(C.getMyCharPtr(resp), C.int(C.getMyCharLen(resp)))
	return errors.New(errMsg)
}

func (n *WakuNode) getConnectedPeers() (peer.IDSlice, error) {
	var resp = C.allocResp()
	defer C.freeResp(resp)
	C.cGoWakuGetConnectedPeers(n.wakuCtx, resp)
	if C.getRet(resp) == C.RET_OK {
		peersStr := C.GoStringN(C.getMyCharPtr(resp), C.int(C.getMyCharLen(resp)))
		if peersStr == "" {
			return nil, nil
		}
		// peersStr contains a comma-separated list of peer ids
		itemsPeerIds := strings.Split(peersStr, ",")
		var peers peer.IDSlice
		for _, peerId := range itemsPeerIds {
			id, err := peer.Decode(peerId)
			if err != nil {
				return nil, fmt.Errorf("GetConnectedPeers - decoding peerId: %w", err)
			}
			peers = append(peers, id)
		}
		return peers, nil
	}
	errMsg := C.GoStringN(C.getMyCharPtr(resp), C.int(C.getMyCharLen(resp)))
	return nil, fmt.Errorf("GetConnectedPeers: %s", errMsg)

}

func (n *WakuNode) relaySubscribe(pubsubTopic string) error {
	if pubsubTopic == "" {
		return errors.New("pubsub topic is empty")
	}

	var resp = C.allocResp()
	var cPubsubTopic = C.CString(pubsubTopic)

	defer C.freeResp(resp)
	defer C.free(unsafe.Pointer(cPubsubTopic))

	if n.wakuCtx == nil {
		return errors.New("wakuCtx is nil")
	}

	C.cGoWakuRelaySubscribe(n.wakuCtx, cPubsubTopic, resp)

	if C.getRet(resp) == C.RET_OK {
		return nil
	}

	errMsg := "error WakuRelaySubscribe: " + C.GoStringN(C.getMyCharPtr(resp), C.int(C.getMyCharLen(resp)))
	return errors.New(errMsg)
}

func (n *WakuNode) relayUnsubscribe(pubsubTopic string) error {
	if pubsubTopic == "" {
		return errors.New("pubsub topic is empty")
	}

	var resp = C.allocResp()
	var cPubsubTopic = C.CString(pubsubTopic)

	defer C.freeResp(resp)
	defer C.free(unsafe.Pointer(cPubsubTopic))

	if n.wakuCtx == nil {
		return errors.New("wakuCtx is nil")
	}

	C.cGoWakuRelayUnsubscribe(n.wakuCtx, cPubsubTopic, resp)

	if C.getRet(resp) == C.RET_OK {
		return nil
	}

	errMsg := "error WakuRelayUnsubscribe: " + C.GoStringN(C.getMyCharPtr(resp), C.int(C.getMyCharLen(resp)))
	return errors.New(errMsg)
}

func (n *WakuNode) peerExchangeRequest(numPeers uint64) (uint64, error) {
	var resp = C.allocResp()
	defer C.freeResp(resp)

	C.cGoWakuPeerExchangeQuery(n.wakuCtx, C.uint64_t(numPeers), resp)
	if C.getRet(resp) == C.RET_OK {
		numRecvPeersStr := C.GoStringN(C.getMyCharPtr(resp), C.int(C.getMyCharLen(resp)))
		numRecvPeers, err := strconv.ParseUint(numRecvPeersStr, 10, 64)
		if err != nil {
			return 0, err
		}
		return numRecvPeers, nil
	}

	errMsg := C.GoStringN(C.getMyCharPtr(resp), C.int(C.getMyCharLen(resp)))
	return 0, errors.New(errMsg)
}

func (n *WakuNode) startDiscV5() error {
	var resp = C.allocResp()
	defer C.freeResp(resp)
	C.cGoWakuStartDiscV5(n.wakuCtx, resp)

	if C.getRet(resp) == C.RET_OK {
		return nil
	}
	errMsg := "error WakuStartDiscV5: " + C.GoStringN(C.getMyCharPtr(resp), C.int(C.getMyCharLen(resp)))
	return errors.New(errMsg)
}

func (n *WakuNode) stopDiscV5() error {
	var resp = C.allocResp()
	defer C.freeResp(resp)
	C.cGoWakuStopDiscV5(n.wakuCtx, resp)

	if C.getRet(resp) == C.RET_OK {
		return nil
	}
	errMsg := "error WakuStopDiscV5: " + C.GoStringN(C.getMyCharPtr(resp), C.int(C.getMyCharLen(resp)))
	return errors.New(errMsg)
}

func (n *WakuNode) version() (string, error) {
	var resp = C.allocResp()
	defer C.freeResp(resp)

	C.cGoWakuVersion(n.wakuCtx, resp)

	if C.getRet(resp) == C.RET_OK {
		var version = C.GoStringN(C.getMyCharPtr(resp), C.int(C.getMyCharLen(resp)))
		return version, nil
	}

	errMsg := "error WakuVersion: " +
		C.GoStringN(C.getMyCharPtr(resp), C.int(C.getMyCharLen(resp)))
	return "", errors.New(errMsg)
}

func (n *WakuNode) storeQuery(storeQueryRequest storeQueryRequest) (any, error) {
	// TODO: extract timeout from context
	timeoutMs := time.Minute.Milliseconds()

	b, err := json.Marshal(storeQueryRequest.storeRequest)
	if err != nil {
		return nil, err
	}

	addrs := make([]string, len(storeQueryRequest.peerInfo.Addrs))
	for i, addr := range utils.EncapsulatePeerID(storeQueryRequest.peerInfo.ID, storeQueryRequest.peerInfo.Addrs...) {
		addrs[i] = addr.String()
	}

	var cJsonQuery = C.CString(string(b))
	var cPeerAddr = C.CString(strings.Join(addrs, ","))
	var resp = C.allocResp()

	defer C.free(unsafe.Pointer(cJsonQuery))
	defer C.free(unsafe.Pointer(cPeerAddr))
	defer C.freeResp(resp)

	C.cGoWakuStoreQuery(n.wakuCtx, cJsonQuery, cPeerAddr, C.int(timeoutMs), resp)
	if C.getRet(resp) == C.RET_OK {
		jsonResponseStr := C.GoStringN(C.getMyCharPtr(resp), C.int(C.getMyCharLen(resp)))
		storeQueryResponse := &storepb.StoreQueryResponse{}
		err = json.Unmarshal([]byte(jsonResponseStr), storeQueryResponse)
		if err != nil {
			return nil, err
		}
		return storeQueryResponse, nil
	}
	errMsg := "error WakuStoreQuery: " +
		C.GoStringN(C.getMyCharPtr(resp), C.int(C.getMyCharLen(resp)))
	return nil, errors.New(errMsg)
}

func (n *WakuNode) relayPublish(relayPublishRequest relayPublishRequest) (pb.MessageHash, error) {
	// TODO: extract timeout from context
	timeoutMs := time.Minute.Milliseconds()

	jsonMsg, err := json.Marshal(relayPublishRequest.message)
	if err != nil {
		return pb.MessageHash{}, err
	}

	var cPubsubTopic = C.CString(relayPublishRequest.pubsubTopic)
	var msg = C.CString(string(jsonMsg))
	var resp = C.allocResp()

	defer C.freeResp(resp)
	defer C.free(unsafe.Pointer(cPubsubTopic))
	defer C.free(unsafe.Pointer(msg))

	C.cGoWakuRelayPublish(n.wakuCtx, cPubsubTopic, msg, C.int(timeoutMs), resp)
	if C.getRet(resp) == C.RET_OK {
		msgHash := C.GoStringN(C.getMyCharPtr(resp), C.int(C.getMyCharLen(resp)))
		msgHashBytes, err := hexutil.Decode(msgHash)
		if err != nil {
			return pb.MessageHash{}, err
		}
		return pb.ToMessageHash(msgHashBytes), nil
	}
	errMsg := "WakuRelayPublish: " + C.GoStringN(C.getMyCharPtr(resp), C.int(C.getMyCharLen(resp)))
	return pb.MessageHash{}, errors.New(errMsg)
}

func (n *WakuNode) dnsDiscovery(dnsDiscRequest dnsDiscoveryRequest) ([]multiaddr.Multiaddr, error) {
	var resp = C.allocResp()
	var cEnrTree = C.CString(dnsDiscRequest.enrTreeUrl)
	var cDnsServer = C.CString(dnsDiscRequest.nameDnsServer)
	defer C.freeResp(resp)
	defer C.free(unsafe.Pointer(cEnrTree))
	defer C.free(unsafe.Pointer(cDnsServer))
	// TODO: extract timeout from context
	C.cGoWakuDnsDiscovery(n.wakuCtx, cEnrTree, cDnsServer, C.int(time.Minute.Milliseconds()), resp)
	if C.getRet(resp) == C.RET_OK {
		var addrsRet []multiaddr.Multiaddr
		nodeAddresses := C.GoStringN(C.getMyCharPtr(resp), C.int(C.getMyCharLen(resp)))
		addrss := strings.Split(nodeAddresses, ",")
		for _, addr := range addrss {
			addr, err := multiaddr.NewMultiaddr(addr)
			if err != nil {
				return nil, err
			}
			addrsRet = append(addrsRet, addr)
		}
		return addrsRet, nil
	}
	errMsg := "error WakuDnsDiscovery: " + C.GoStringN(C.getMyCharPtr(resp), C.int(C.getMyCharLen(resp)))
	return nil, errors.New(errMsg)
}

func (n *WakuNode) pingPeer(request pingRequest) (time.Duration, error) {
	peerInfo := request.peerInfo

	addrs := make([]string, len(peerInfo.Addrs))
	for i, addr := range utils.EncapsulatePeerID(peerInfo.ID, peerInfo.Addrs...) {
		addrs[i] = addr.String()
	}

	var resp = C.allocResp()
	var cPeerId = C.CString(strings.Join(addrs, ","))
	defer C.freeResp(resp)
	defer C.free(unsafe.Pointer(cPeerId))

	// TODO: extract timeout from ctx
	C.cGoWakuPingPeer(n.wakuCtx, cPeerId, C.int(time.Minute.Milliseconds()), resp)
	if C.getRet(resp) == C.RET_OK {
		rttStr := C.GoStringN(C.getMyCharPtr(resp), C.int(C.getMyCharLen(resp)))
		rttInt, err := strconv.ParseInt(rttStr, 10, 64)
		if err != nil {
			return 0, err
		}
		return time.Duration(rttInt), nil
	}

	errMsg := C.GoStringN(C.getMyCharPtr(resp), C.int(C.getMyCharLen(resp)))
	return 0, fmt.Errorf("PingPeer: %s", errMsg)
}

func (n *WakuNode) start() error {
	var resp = C.allocResp()
	defer C.freeResp(resp)

	C.cGoWakuStart(n.wakuCtx, resp)

	if C.getRet(resp) == C.RET_OK {
		return nil
	}

	errMsg := "error WakuStart: " + C.GoStringN(C.getMyCharPtr(resp), C.int(C.getMyCharLen(resp)))
	return errors.New(errMsg)
}

func (n *WakuNode) stop() error {
	var resp = C.allocResp()
	defer C.freeResp(resp)

	C.cGoWakuStop(n.wakuCtx, resp)

	if C.getRet(resp) == C.RET_OK {
		return nil
	}

	errMsg := "error WakuStop: " + C.GoStringN(C.getMyCharPtr(resp), C.int(C.getMyCharLen(resp)))
	return errors.New(errMsg)
}

func (n *WakuNode) destroy() error {
	var resp = C.allocResp()
	defer C.freeResp(resp)

	C.cGoWakuDestroy(n.wakuCtx, resp)

	if C.getRet(resp) == C.RET_OK {
		return nil
	}

	errMsg := "error WakuDestroy: " + C.GoStringN(C.getMyCharPtr(resp), C.int(C.getMyCharLen(resp)))
	return errors.New(errMsg)
}

func (n *WakuNode) peerID() (peer.ID, error) {
	var resp = C.allocResp()
	defer C.freeResp(resp)

	C.cGoWakuGetMyPeerId(n.wakuCtx, resp)

	if C.getRet(resp) == C.RET_OK {
		peerIdStr := C.GoStringN(C.getMyCharPtr(resp), C.int(C.getMyCharLen(resp)))
		id, err := peer.Decode(peerIdStr)
		if err != nil {
			errMsg := "WakuGetMyPeerId - decoding peerId: %w"
			return "", fmt.Errorf(errMsg, err)
		}
		return id, nil
	}
	errMsg := C.GoStringN(C.getMyCharPtr(resp), C.int(C.getMyCharLen(resp)))
	return "", errors.New(errMsg)
}

func (n *WakuNode) connect(connReq connectRequest) error {
	var resp = C.allocResp()
	var cPeerMultiAddr = C.CString(connReq.addr.String())
	defer C.freeResp(resp)
	defer C.free(unsafe.Pointer(cPeerMultiAddr))

	// TODO: extract timeout from ctx
	C.cGoWakuConnect(n.wakuCtx, cPeerMultiAddr, C.int(time.Minute.Milliseconds()), resp)

	if C.getRet(resp) == C.RET_OK {
		return nil
	}
	errMsg := "error WakuConnect: " +
		C.GoStringN(C.getMyCharPtr(resp), C.int(C.getMyCharLen(resp)))
	return errors.New(errMsg)
}

func (n *WakuNode) dialPeerById(dialPeerByIDReq dialPeerByIDRequest) error {
	var resp = C.allocResp()
	var cPeerId = C.CString(dialPeerByIDReq.peerID.String())
	var cProtocol = C.CString(string(dialPeerByIDReq.protocol))
	defer C.freeResp(resp)
	defer C.free(unsafe.Pointer(cPeerId))
	defer C.free(unsafe.Pointer(cProtocol))

	// TODO: extract timeout from ctx
	C.cGoWakuDialPeerById(n.wakuCtx, cPeerId, cProtocol, C.int(time.Minute.Milliseconds()), resp)

	if C.getRet(resp) == C.RET_OK {
		return nil
	}
	errMsg := "error DialPeerById: " +
		C.GoStringN(C.getMyCharPtr(resp), C.int(C.getMyCharLen(resp)))
	return errors.New(errMsg)
}

func (n *WakuNode) listenAddresses() ([]multiaddr.Multiaddr, error) {
	var resp = C.allocResp()
	defer C.freeResp(resp)
	C.cGoWakuListenAddresses(n.wakuCtx, resp)

	if C.getRet(resp) == C.RET_OK {
		var addrsRet []multiaddr.Multiaddr
		listenAddresses := C.GoStringN(C.getMyCharPtr(resp), C.int(C.getMyCharLen(resp)))
		addrss := strings.Split(listenAddresses, ",")
		for _, addr := range addrss {
			addr, err := multiaddr.NewMultiaddr(addr)
			if err != nil {
				return nil, err
			}
			addrsRet = append(addrsRet, addr)
		}
		return addrsRet, nil
	}
	errMsg := "error WakuListenAddresses: " + C.GoStringN(C.getMyCharPtr(resp), C.int(C.getMyCharLen(resp)))
	return nil, errors.New(errMsg)
}

func (n *WakuNode) enr() (*enode.Node, error) {
	var resp = C.allocResp()
	defer C.freeResp(resp)
	C.cGoWakuGetMyENR(n.wakuCtx, resp)

	if C.getRet(resp) == C.RET_OK {
		enrStr := C.GoStringN(C.getMyCharPtr(resp), C.int(C.getMyCharLen(resp)))
		n, err := enode.Parse(enode.ValidSchemes, enrStr)
		if err != nil {
			return nil, err
		}
		return n, nil
	}
	errMsg := "error WakuGetMyENR: " +
		C.GoStringN(C.getMyCharPtr(resp), C.int(C.getMyCharLen(resp)))
	return nil, errors.New(errMsg)
}

func (n *WakuNode) listPeersInMesh(pubsubTopic string) (int, error) {
	var resp = C.allocResp()
	var cPubsubTopic = C.CString(pubsubTopic)
	defer C.freeResp(resp)
	defer C.free(unsafe.Pointer(cPubsubTopic))

	C.cGoWakuListPeersInMesh(n.wakuCtx, cPubsubTopic, resp)

	if C.getRet(resp) == C.RET_OK {
		numPeersStr := C.GoStringN(C.getMyCharPtr(resp), C.int(C.getMyCharLen(resp)))
		numPeers, err := strconv.Atoi(numPeersStr)
		if err != nil {
			errMsg := "ListPeersInMesh - error converting string to int: " + err.Error()
			return 0, errors.New(errMsg)
		}
		return numPeers, nil
	}
	errMsg := "error ListPeersInMesh: " +
		C.GoStringN(C.getMyCharPtr(resp), C.int(C.getMyCharLen(resp)))
	return 0, errors.New(errMsg)
}

func (n *WakuNode) getPeerIDsFromPeerStore() (peer.IDSlice, error) {
	var resp = C.allocResp()
	defer C.freeResp(resp)
	C.cGoWakuGetPeerIdsFromPeerStore(n.wakuCtx, resp)

	if C.getRet(resp) == C.RET_OK {
		peersStr := C.GoStringN(C.getMyCharPtr(resp), C.int(C.getMyCharLen(resp)))
		if peersStr == "" {
			return nil, nil
		}
		// peersStr contains a comma-separated list of peer ids
		itemsPeerIds := strings.Split(peersStr, ",")

		var peers peer.IDSlice
		for _, peerId := range itemsPeerIds {
			id, err := peer.Decode(peerId)
			if err != nil {
				return nil, fmt.Errorf("GetPeerIdsFromPeerStore - decoding peerId: %w", err)
			}
			peers = append(peers, id)
		}
		return peers, nil
	}
	errMsg := C.GoStringN(C.getMyCharPtr(resp), C.int(C.getMyCharLen(resp)))
	return nil, fmt.Errorf("GetPeerIdsFromPeerStore: %s", errMsg)
}

func (n *WakuNode) getPeerIDsByProtocol(protocolID libp2pproto.ID) (peer.IDSlice, error) {
	var resp = C.allocResp()
	var cProtocol = C.CString(string(protocolID))
	defer C.freeResp(resp)
	defer C.free(unsafe.Pointer(cProtocol))

	C.cGoWakuGetPeerIdsByProtocol(n.wakuCtx, cProtocol, resp)

	if C.getRet(resp) == C.RET_OK {
		peersStr := C.GoStringN(C.getMyCharPtr(resp), C.int(C.getMyCharLen(resp)))
		if peersStr == "" {
			return nil, nil
		}
		// peersStr contains a comma-separated list of peer ids
		itemsPeerIds := strings.Split(peersStr, ",")

		var peers peer.IDSlice
		for _, p := range itemsPeerIds {
			id, err := peer.Decode(p)
			if err != nil {
				return nil, fmt.Errorf("GetPeerIdsByProtocol - decoding peerId: %w", err)
			}
			peers = append(peers, id)
		}
		return peers, nil
	}
	errMsg := "error GetPeerIdsByProtocol: " +
		C.GoStringN(C.getMyCharPtr(resp), C.int(C.getMyCharLen(resp)))
	return nil, fmt.Errorf("GetPeerIdsByProtocol: %s", errMsg)
}

func (n *WakuNode) newNode(config *WakuConfig) error {
	jsonConfig, err := json.Marshal(config)
	if err != nil {
		return err
	}

	var cJsonConfig = C.CString(string(jsonConfig))
	var resp = C.allocResp()

	defer C.free(unsafe.Pointer(cJsonConfig))
	defer C.freeResp(resp)

	if C.getRet(resp) != C.RET_OK {
		errMsg := "error wakuNew: " + C.GoStringN(C.getMyCharPtr(resp), C.int(C.getMyCharLen(resp)))
		return errors.New(errMsg)
	}

	wakuCtx := C.cGoWakuNew(cJsonConfig, resp)
	n.wakuCtx = unsafe.Pointer(wakuCtx)

	// Notice that the events for self node are handled by the 'MyEventCallback' method
	C.cGoWakuSetEventCallback(n.wakuCtx)

	return nil
}

func (n *WakuNode) dialPeer(dialPeerReq dialPeerRequest) error {
	var resp = C.allocResp()
	var cPeerMultiAddr = C.CString(dialPeerReq.peerAddr.String())
	var cProtocol = C.CString(string(dialPeerReq.protocol))
	defer C.freeResp(resp)
	defer C.free(unsafe.Pointer(cPeerMultiAddr))
	defer C.free(unsafe.Pointer(cProtocol))
	// TODO: extract timeout from context
	C.cGoWakuDialPeer(n.wakuCtx, cPeerMultiAddr, cProtocol, C.int(requestTimeout.Milliseconds()), resp)
	if C.getRet(resp) == C.RET_OK {
		return nil
	}
	errMsg := "error DialPeer: " + C.GoStringN(C.getMyCharPtr(resp), C.int(C.getMyCharLen(resp)))
	return errors.New(errMsg)
}

type pingRequest struct {
	ctx      context.Context
	peerInfo peer.AddrInfo
}

func (n *WakuNode) PingPeer(ctx context.Context, info peer.AddrInfo) (time.Duration, error) {
	response, err := n.postTask(requestTypePing, pingRequest{
		ctx:      ctx,
		peerInfo: info,
	})
	if err != nil {
		return 0, err
	}
	return response.(time.Duration), nil
}

func (n *WakuNode) Start() error {
	_, err := n.postTask(requestTypeStart, nil)
	return err
}

type relayPublishRequest struct {
	ctx         context.Context
	pubsubTopic string
	message     *pb.WakuMessage
}

func (n *WakuNode) RelayPublish(ctx context.Context, message *pb.WakuMessage, pubsubTopic string) (pb.MessageHash, error) {
	response, err := n.postTask(requestTypeRelayPublish, relayPublishRequest{
		ctx:         ctx,
		pubsubTopic: pubsubTopic,
		message:     message,
	})
	if err != nil {
		return pb.MessageHash{}, err
	}
	return response.(pb.MessageHash), nil
}

type storeQueryRequest struct {
	ctx          context.Context
	storeRequest *storepb.StoreQueryRequest
	peerInfo     peer.AddrInfo
}

func (n *WakuNode) StoreQuery(ctx context.Context, storeRequest *storepb.StoreQueryRequest, peerInfo peer.AddrInfo) (*storepb.StoreQueryResponse, error) {
	response, err := n.postTask(requestTypeStoreQuery, storeQueryRequest{
		ctx:          ctx,
		peerInfo:     peerInfo,
		storeRequest: storeRequest,
	})
	if err != nil {
		return nil, err
	}
	return response.(*storepb.StoreQueryResponse), nil
}

func (n *WakuNode) PeerID() (peer.ID, error) {
	response, err := n.postTask(requestTypePeerID, nil)
	if err != nil {
		return "", err
	}
	return response.(peer.ID), nil
}

func (n *WakuNode) Stop() error {
	_, err := n.postTask(requestTypeStop, nil)
	return err
}

func (n *WakuNode) Destroy() error {
	_, err := n.postTask(requestTypeDestroy, nil)
	return err
}

func (n *WakuNode) StartDiscV5() error {
	_, err := n.postTask(requestTypeStartDiscV5, nil)
	return err
}

func (n *WakuNode) StopDiscV5() error {
	_, err := n.postTask(requestTypeStopDiscV5, nil)
	return err
}

func (n *WakuNode) Version() (string, error) {
	response, err := n.postTask(requestTypeVersion, nil)
	if err != nil {
		return "", err
	}
	return response.(string), nil
}

func (n *WakuNode) RelaySubscribe(pubsubTopic string) error {
	_, err := n.postTask(requestTypeRelaySubscribe, pubsubTopic)
	return err
}

func (n *WakuNode) RelayUnsubscribe(pubsubTopic string) error {
	_, err := n.postTask(requestTypeRelayUnsubscribe, pubsubTopic)
	return err
}

func (n *WakuNode) PeerExchangeRequest(numPeers uint64) (uint64, error) {
	response, err := n.postTask(requestTypePeerExchangeRequest, numPeers)
	if err != nil {
		return 0, err
	}
	return response.(uint64), nil
}

type connectRequest struct {
	ctx  context.Context
	addr multiaddr.Multiaddr
}

func (n *WakuNode) Connect(ctx context.Context, addr multiaddr.Multiaddr) error {
	_, err := n.postTask(requestTypeConnect, connectRequest{
		ctx:  ctx,
		addr: addr,
	})
	return err
}

type dialPeerByIDRequest struct {
	ctx      context.Context
	peerID   peer.ID
	protocol libp2pproto.ID
}

func (n *WakuNode) DialPeerByID(ctx context.Context, peerID peer.ID, protocol libp2pproto.ID) error {
	_, err := n.postTask(requestTypeDialPeerByID, dialPeerByIDRequest{
		ctx:      ctx,
		peerID:   peerID,
		protocol: protocol,
	})
	return err
}

func (n *WakuNode) ListenAddresses() ([]multiaddr.Multiaddr, error) {
	response, err := n.postTask(requestTypeListenAddresses, nil)
	if err != nil {
		return nil, err
	}
	return response.([]multiaddr.Multiaddr), nil
}

func (n *WakuNode) ENR() (*enode.Node, error) {
	response, err := n.postTask(requestTypeENR, nil)
	if err != nil {
		return nil, err
	}
	return response.(*enode.Node), nil
}

func (n *WakuNode) ListPeersInMesh(pubsubTopic string) (int, error) {
	response, err := n.postTask(requestTypeListPeersInMesh, pubsubTopic)
	if err != nil {
		return 0, err
	}
	return response.(int), nil
}

func (n *WakuNode) GetConnectedPeers() (peer.IDSlice, error) {
	response, err := n.postTask(requestTypeGetConnectedPeers, nil)
	if err != nil {
		return nil, err
	}
	return response.(peer.IDSlice), nil
}

func (n *WakuNode) GetNumConnectedPeers() (int, error) {
	peers, err := n.GetConnectedPeers()
	if err != nil {
		return 0, err
	}
	return len(peers), nil
}

func (n *WakuNode) GetPeerIDsFromPeerStore() (peer.IDSlice, error) {
	response, err := n.postTask(requestTypeGetPeerIDsFromPeerStore, nil)
	if err != nil {
		return nil, err
	}
	return response.(peer.IDSlice), nil
}

func (n *WakuNode) GetPeerIDsByProtocol(protocol libp2pproto.ID) (peer.IDSlice, error) {
	response, err := n.postTask(requestTypeGetPeerIDsByProtocol, protocol)
	if err != nil {
		return nil, err
	}
	return response.(peer.IDSlice), nil
}

func (n *WakuNode) DisconnectPeerByID(peerID peer.ID) error {
	_, err := n.postTask(requestTypeDisconnectPeerByID, peerID)
	return err
}

type dnsDiscoveryRequest struct {
	ctx           context.Context
	enrTreeUrl    string
	nameDnsServer string
}

func (n *WakuNode) DnsDiscovery(ctx context.Context, enrTreeUrl string, nameDnsServer string) ([]multiaddr.Multiaddr, error) {
	response, err := n.postTask(requestTypeDnsDiscovery, dnsDiscoveryRequest{
		ctx:           ctx,
		enrTreeUrl:    enrTreeUrl,
		nameDnsServer: nameDnsServer,
	})
	if err != nil {
		return nil, err
	}
	return response.([]multiaddr.Multiaddr), nil
}

type dialPeerRequest struct {
	ctx      context.Context
	peerAddr multiaddr.Multiaddr
	protocol libp2pproto.ID
}

func (n *WakuNode) DialPeer(ctx context.Context, peerAddr multiaddr.Multiaddr, protocol libp2pproto.ID) error {
	_, err := n.postTask(requestTypeDialPeer, dialPeerRequest{
		ctx:      ctx,
		peerAddr: peerAddr,
		protocol: protocol,
	})
	return err
}

func (n *WakuNode) GetNumConnectedRelayPeers(paramPubsubTopic ...string) (int, error) {
	response, err := n.postTask(requestTypeGetNumConnectedRelayPeers, paramPubsubTopic)
	if err != nil {
		return 0, err
	}
	return response.(int), nil
}
