//go:build dev
// +build dev

package digest

import (
	"net/http"
	"runtime"
)

var (
	HandlerPathResource       = `/resource`
	HandlerPathResourceProm   = `/resource/prom`
	HandlerPathPProfProfile   = `/pprof/profile`
	HandlerPathPProfGoroutine = `/pprof/goroutine`
	HandlerPathPProfHeap      = `/pprof/heap`
	HandlerPathPProfBlock     = `/pprof/block`
	HandlerPathPProfMutex     = `/pprof/mutex`
	HandlerPathPProfAllocs    = `/pprof/allocs`
)

func (hd *Handlers) setHandlers(digest bool) {
	runtime.SetBlockProfileRate(1)
	post := 5
	postQueue := 10000
	get := 1000

	_ = hd.setHandler(HandlerPathSend, hd.handleSend, false, post, post).
		Methods(http.MethodOptions, http.MethodPost)
	_ = hd.setHandler(HandlerPathQueueSend, hd.handleQueueSend, false, postQueue, postQueue).
		Methods(http.MethodOptions, http.MethodPost)
	_ = hd.setHandler(HandlerPathNodeInfo, hd.handleNodeInfo, true, get, get).
		Methods(http.MethodOptions, "GET")
	_ = hd.setHandler(HandlerPathNodeMetric, hd.handleNodeMetric, true, get, get).
		Methods(http.MethodOptions, "GET")
	_ = hd.setHandler(HandlerPathNodeInfoProm, hd.handleNodeInfoProm, false, get, get).
		Methods(http.MethodOptions, "GET")
	_ = hd.setHandler(HandlerPathNodeMetricProm, hd.handleNodeMetricProm, false, get, get).
		Methods(http.MethodOptions, "GET")
	_ = hd.setHandler(HandlerPathResource, hd.handleResource, true, get, get).
		Methods(http.MethodOptions, "GET")
	_ = hd.setHandler(HandlerPathResourceProm, hd.handleResourceProm, false, get, get).
		Methods(http.MethodOptions, "GET")
	_ = hd.setHandler(HandlerPathPProfProfile, hd.handlePProfProfile, true, get, get).
		Methods(http.MethodOptions, "GET")
	_ = hd.setHandler(HandlerPathPProfHeap, hd.handlePProfHeap, true, get, get).
		Methods(http.MethodOptions, "GET")
	_ = hd.setHandler(HandlerPathPProfAllocs, hd.handlePProfAllocs, true, get, get).
		Methods(http.MethodOptions, "GET")
	_ = hd.setHandler(HandlerPathPProfGoroutine, hd.handlePProfGoroutine, true, get, get).
		Methods(http.MethodOptions, "GET")
	_ = hd.setHandler(HandlerPathPProfBlock, hd.handlePProfBlock, true, get, get).
		Methods(http.MethodOptions, "GET")

	if digest {
		_ = hd.setHandler(HandlerPathCurrencies, hd.handleCurrencies, true, get, get).
			Methods(http.MethodOptions, "GET")
		_ = hd.setHandler(HandlerPathCurrency, hd.handleCurrency, true, get, get).
			Methods(http.MethodOptions, "GET")
		_ = hd.setHandler(HandlerPathManifests, hd.handleManifests, true, get, get).
			Methods(http.MethodOptions, "GET")
		_ = hd.setHandler(HandlerPathOperations, hd.handleOperations, true, get, get).
			Methods(http.MethodOptions, "GET")
		_ = hd.setHandler(HandlerPathOperationsByHash, hd.handleOperationsByHash, true, get, get).
			Methods(http.MethodOptions, "GET")
		_ = hd.setHandler(HandlerPathOperation, hd.handleOperation, true, get, get).
			Methods(http.MethodOptions, "GET")
		_ = hd.setHandler(HandlerPathOperationsByHeight, hd.handleOperationsByHeight, true, get, get).
			Methods(http.MethodOptions, "GET")
		_ = hd.setHandler(HandlerPathManifestByHeight, hd.handleManifestByHeight, true, get, get).
			Methods(http.MethodOptions, "GET")
		_ = hd.setHandler(HandlerPathManifestByHash, hd.handleManifestByHash, true, get, get).
			Methods(http.MethodOptions, "GET")
		_ = hd.setHandler(HandlerPathBlockByHeight, hd.handleBlock, true, get, get).
			Methods(http.MethodOptions, "GET")
		_ = hd.setHandler(HandlerPathBlockByHash, hd.handleBlock, true, get, get).
			Methods(http.MethodOptions, "GET")
		_ = hd.setHandler(HandlerPathAccount, hd.handleAccount, true, get, get).
			Methods(http.MethodOptions, "GET")
		_ = hd.setHandler(HandlerPathAccountOperations, hd.handleAccountOperations, true, get, get).
			Methods(http.MethodOptions, "GET")
		_ = hd.setHandler(HandlerPathAccounts, hd.handleAccounts, true, get, get).
			Methods(http.MethodOptions, "GET")
		_ = hd.setHandler(HandlerPathDIDData, hd.handleDIDData, true, get, get).
			Methods(http.MethodOptions, "GET")
		_ = hd.setHandler(HandlerPathDIDDesign, hd.handleDIDDesign, true, get, get).
			Methods(http.MethodOptions, "GET")
		_ = hd.setHandler(HandlerPathDIDDocument, hd.handleDIDDocument, true, get, get).
			Methods(http.MethodOptions, "GET")
	}
}
