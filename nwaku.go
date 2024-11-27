//go:build use_nwaku
// +build use_nwaku

package wakuv2

/*
	#cgo LDFLAGS: -L../third_party/nwaku/build/ -lnegentropy -lwaku
	#cgo LDFLAGS: -L../third_party/nwaku -Wl,-rpath,../third_party/nwaku/build/

	#include "./third_party/nwaku/library/libwaku.h"
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
	"errors"
	"runtime"
	"unsafe"

	"github.com/libp2p/go-libp2p/core/peer"
	"go.uber.org/zap"
)

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
		defer gocommon.LogOnPanic()

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

	logger.Info("starting wakuv2 with config", zap.Any("nwakuCfg", nwakuCfg), zap.Any("wakuCfg", cfg))

	wakunode, err := newWakuNode(ctx, nwakuCfg)
	if err != nil {
		cancel()
		return nil, err
	}

	return &Waku{
		node:    wakunode,
		wakuCfg: nwakuCfg,
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
