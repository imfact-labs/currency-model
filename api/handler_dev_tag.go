//go:build dev
// +build dev

package api

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

func SetHandlers(hd *Handlers, digest bool) {
	runtime.SetBlockProfileRate(1)
	post := 5
	postQueue := 10000
	get := 1000

	_ = hd.SetHandler(HandlerPathSend, HandleSend, false, post, post).
		Methods(http.MethodOptions, http.MethodPost)
	_ = hd.SetHandler(HandlerPathQueueSend, HandleQueueSend, false, postQueue, postQueue).
		Methods(http.MethodOptions, http.MethodPost)
	_ = hd.SetHandler(HandlerPathNodeInfo, HandleNodeInfo, true, get, get).
		Methods(http.MethodOptions, "GET")
	_ = hd.SetHandler(HandlerPathNodeMetric, HandleNodeMetric, true, get, get).
		Methods(http.MethodOptions, "GET")
	_ = hd.SetHandler(HandlerPathNodeInfoProm, HandleNodeInfoProm, false, get, get).
		Methods(http.MethodOptions, "GET")
	_ = hd.SetHandler(HandlerPathNodeMetricProm, HandleNodeMetricProm, false, get, get).
		Methods(http.MethodOptions, "GET")
	_ = hd.SetHandler(HandlerPathResource, HandleResource, true, get, get).
		Methods(http.MethodOptions, "GET")
	_ = hd.SetHandler(HandlerPathResourceProm, HandleResourceProm, false, get, get).
		Methods(http.MethodOptions, "GET")
	_ = hd.SetHandler(HandlerPathPProfProfile, HandlePProfProfile, true, get, get).
		Methods(http.MethodOptions, "GET")
	_ = hd.SetHandler(HandlerPathPProfHeap, HandlePProfHeap, true, get, get).
		Methods(http.MethodOptions, "GET")
	_ = hd.SetHandler(HandlerPathPProfAllocs, HandlePProfAllocs, true, get, get).
		Methods(http.MethodOptions, "GET")
	_ = hd.SetHandler(HandlerPathPProfGoroutine, HandlePProfGoroutine, true, get, get).
		Methods(http.MethodOptions, "GET")
	_ = hd.SetHandler(HandlerPathPProfBlock, HandlePProfBlock, true, get, get).
		Methods(http.MethodOptions, "GET")

	if digest {
		_ = hd.SetHandler(HandlerPathCurrencies, HandleCurrencies, true, get, get).
			Methods(http.MethodOptions, "GET")
		_ = hd.SetHandler(HandlerPathCurrency, HandleCurrency, true, get, get).
			Methods(http.MethodOptions, "GET")
		_ = hd.SetHandler(HandlerPathManifests, HandleManifests, true, get, get).
			Methods(http.MethodOptions, "GET")
		_ = hd.SetHandler(HandlerPathOperations, HandleOperations, true, get, get).
			Methods(http.MethodOptions, "GET")
		_ = hd.SetHandler(HandlerPathOperationsByHash, HandleOperationsByHash, true, get, get).
			Methods(http.MethodOptions, "GET")
		_ = hd.SetHandler(HandlerPathOperation, HandleOperation, true, get, get).
			Methods(http.MethodOptions, "GET")
		_ = hd.SetHandler(HandlerPathOperationsByHeight, HandleOperationsByHeight, true, get, get).
			Methods(http.MethodOptions, "GET")
		_ = hd.SetHandler(HandlerPathManifestByHeight, HandleManifestByHeight, true, get, get).
			Methods(http.MethodOptions, "GET")
		_ = hd.SetHandler(HandlerPathManifestByHash, HandleManifestByHash, true, get, get).
			Methods(http.MethodOptions, "GET")
		_ = hd.SetHandler(HandlerPathBlockByHeight, HandleBlock, true, get, get).
			Methods(http.MethodOptions, "GET")
		_ = hd.SetHandler(HandlerPathBlockByHash, HandleBlock, true, get, get).
			Methods(http.MethodOptions, "GET")
		_ = hd.SetHandler(HandlerPathAccount, HandleAccount, true, get, get).
			Methods(http.MethodOptions, "GET")
		_ = hd.SetHandler(HandlerPathAccountOperations, HandleAccountOperations, true, get, get).
			Methods(http.MethodOptions, "GET")
		_ = hd.SetHandler(HandlerPathAccounts, HandleAccounts, true, get, get).
			Methods(http.MethodOptions, "GET")
		_ = hd.SetHandler(HandlerPathDIDData, HandleDIDData, true, get, get).
			Methods(http.MethodOptions, "GET")
		_ = hd.SetHandler(HandlerPathDIDDesign, HandleDIDDesign, true, get, get).
			Methods(http.MethodOptions, "GET")
		_ = hd.SetHandler(HandlerPathDIDDocument, HandleDIDDocument, true, get, get).
			Methods(http.MethodOptions, "GET")
	}
}
