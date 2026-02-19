package api

import (
	"net/http"

	"github.com/imfact-labs/currency-model/digest"
	"github.com/imfact-labs/currency-model/types"
	"github.com/imfact-labs/mitum2/base"
	"github.com/pkg/errors"
)

var (
	HandlerPathDIDDesign   = `/did-registry/{contract:(?i)` + types.REStringAddressString + `}`
	HandlerPathDIDData     = `/did-registry/{contract:(?i)` + types.REStringAddressString + `}/did/{method_specific_id:` + types.ReSpecialCh + `}`
	HandlerPathDIDDocument = `/did-registry/{contract:(?i)` + types.REStringAddressString + `}/document`
)

func HandleDIDDesign(hd *Handlers, w http.ResponseWriter, r *http.Request) {
	cacheKey := CacheKeyPath(r)
	if err := LoadFromCache(hd.cache, cacheKey, w); err == nil {
		return
	}

	contract, err, status := ParseRequest(w, r, "contract")
	if err != nil {
		HTTP2ProblemWithError(w, err, status)
		return
	}

	if v, err, shared := hd.rg.Do(cacheKey, func() (interface{}, error) {
		return handleDIDDesignInGroup(hd, contract)
	}); err != nil {
		HTTP2HandleError(w, err)
	} else {
		HTTP2WriteHalBytes(hd.enc, w, v.([]byte), http.StatusOK)

		if !shared {
			HTTP2WriteCache(w, cacheKey, hd.expireLongLived)
		}
	}
}

func handleDIDDesignInGroup(hd *Handlers, contract string) ([]byte, error) {
	var de types.Design
	var st base.State

	de, st, err := digest.DIDDesign(hd.database, contract)
	if err != nil {
		return nil, err
	}

	i, err := buildDIDDesign(hd, contract, de, st)
	if err != nil {
		return nil, err
	}
	return hd.enc.Marshal(i)
}

func buildDIDDesign(hd *Handlers, contract string, de types.Design, st base.State) (Hal, error) {
	h, err := hd.CombineURL(HandlerPathDIDDesign, "contract", contract)
	if err != nil {
		return nil, err
	}

	var hal Hal
	hal = NewBaseHal(de, NewHalLink(h, nil))

	h, err = hd.CombineURL(HandlerPathBlockByHeight, "height", st.Height().String())
	if err != nil {
		return nil, err
	}
	hal = hal.AddLink("block", NewHalLink(h, nil))

	for i := range st.Operations() {
		h, err := hd.CombineURL(HandlerPathOperation, "hash", st.Operations()[i].String())
		if err != nil {
			return nil, err
		}
		hal = hal.AddLink("operations", NewHalLink(h, nil))
	}

	return hal, nil
}

func HandleDIDData(hd *Handlers, w http.ResponseWriter, r *http.Request) {
	cacheKey := CacheKeyPath(r)
	if err := LoadFromCache(hd.cache, cacheKey, w); err == nil {
		return
	}

	contract, err, status := ParseRequest(w, r, "contract")
	if err != nil {
		HTTP2ProblemWithError(w, err, status)
		return
	}

	key, err, status := ParseRequest(w, r, "method_specific_id")
	if err != nil {
		HTTP2ProblemWithError(w, err, status)
		return
	}

	if v, err, shared := hd.rg.Do(cacheKey, func() (interface{}, error) {
		return handleDIDDataInGroup(hd, contract, key)
	}); err != nil {
		HTTP2HandleError(w, err)
	} else {
		HTTP2WriteHalBytes(hd.enc, w, v.([]byte), http.StatusOK)

		if !shared {
			HTTP2WriteCache(w, cacheKey, hd.expireLongLived)
		}
	}
}

func handleDIDDataInGroup(hd *Handlers, contract, key string) ([]byte, error) {
	data, st, err := digest.DIDData(hd.database, contract, key)
	if err != nil {
		return nil, err
	}

	i, err := hd.buildDIDDataHal(contract, *data, st)
	if err != nil {
		return nil, err
	}
	return hd.enc.Marshal(i)
}

func (hd *Handlers) buildDIDDataHal(
	contract string, data types.Data, st base.State) (Hal, error) {
	h, err := hd.CombineURL(
		HandlerPathDIDData,
		"contract", contract, "method_specific_id", data.Address().String())
	if err != nil {
		return nil, err
	}

	var hal Hal
	hal = NewBaseHal(data, NewHalLink(h, nil))
	h, err = hd.CombineURL(HandlerPathBlockByHeight, "height", st.Height().String())
	if err != nil {
		return nil, err
	}
	hal = hal.AddLink("block", NewHalLink(h, nil))

	for i := range st.Operations() {
		h, err := hd.CombineURL(HandlerPathOperation, "hash", st.Operations()[i].String())
		if err != nil {
			return nil, err
		}
		hal = hal.AddLink("operations", NewHalLink(h, nil))
	}

	return hal, nil
}

func HandleDIDDocument(hd *Handlers, w http.ResponseWriter, r *http.Request) {
	cacheKey := CacheKeyPath(r)
	if err := LoadFromCache(hd.cache, cacheKey, w); err == nil {
		return
	}

	contract, err, status := ParseRequest(w, r, "contract")
	if err != nil {
		HTTP2ProblemWithError(w, err, status)
		return
	}

	did := ParseStringQuery(r.URL.Query().Get("did"))
	if len(did) < 1 {
		HTTP2ProblemWithError(w, errors.Errorf("invalid DID"), http.StatusBadRequest)
		return
	}

	if v, err, shared := hd.rg.Do(cacheKey, func() (interface{}, error) {
		return handleDIDDocumentInGroup(hd, contract, did)
	}); err != nil {
		HTTP2HandleError(w, err)
	} else {
		HTTP2WriteHalBytes(hd.enc, w, v.([]byte), http.StatusOK)

		if !shared {
			HTTP2WriteCache(w, cacheKey, hd.expireShortLived)
		}
	}
}

func handleDIDDocumentInGroup(hd *Handlers, contract, key string) ([]byte, error) {
	doc, st, err := digest.DIDDocument(hd.database, contract, key)
	if err != nil {
		return nil, err
	}

	i, err := buildDIDDocumentHal(hd, contract, *doc, st)
	if err != nil {
		return nil, err
	}
	return hd.enc.Marshal(i)
}

func buildDIDDocumentHal(
	hd *Handlers, contract string, doc types.DIDDocument, st base.State) (Hal, error) {
	//h, err := hd.CombineURL(
	//	HandlerPathDIDDocument,
	//	"contract", contract)
	//if err != nil {
	//	return nil, err
	//}

	var hal Hal
	hal = NewBaseHal(doc, NewHalLink("", nil))
	h, err := hd.CombineURL(HandlerPathBlockByHeight, "height", st.Height().String())
	if err != nil {
		return nil, err
	}
	hal = hal.AddLink("block", NewHalLink(h, nil))

	for i := range st.Operations() {
		h, err := hd.CombineURL(HandlerPathOperation, "hash", st.Operations()[i].String())
		if err != nil {
			return nil, err
		}
		hal = hal.AddLink("operations", NewHalLink(h, nil))
	}

	return hal, nil
}
