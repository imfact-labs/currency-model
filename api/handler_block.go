package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/imfact-labs/currency-model/digest"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util"
)

var halBlockTemplate = map[string]HalLink{
	"block:{height}":    NewHalLink(HandlerPathBlockByHeight, nil).SetTemplated(),
	"block:{hash}":      NewHalLink(HandlerPathBlockByHeight, nil).SetTemplated(),
	"manifest:{height}": NewHalLink(HandlerPathManifestByHeight, nil).SetTemplated(),
	"manifest:{hash}":   NewHalLink(HandlerPathManifestByHash, nil).SetTemplated(),
}

func HandleBlock(hd *Handlers, w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	cachekey := CacheKeyPath(r)
	if err := LoadFromCache(hd.cache, cachekey, w); err == nil {
		return
	}

	if v, err, shared := hd.rg.Do(cachekey, func() (interface{}, error) {
		return handleBlockInGroup(hd, mux.Vars(r))
	}); err != nil {
		HTTP2HandleError(w, err)
	} else {
		HTTP2WriteHalBytes(hd.enc, w, v.([]byte), http.StatusOK)
		if !shared {
			HTTP2WriteCache(w, cachekey, hd.expireLongLived)
		}
	}
}

func handleBlockInGroup(hd *Handlers, vars map[string]string) ([]byte, error) {
	var hal Hal
	if s, found := vars["height"]; found {
		height, err := parseHeightFromPath(s)
		if err != nil {
			return nil, digest.ErrBadRequest.Errorf("Invalid height found for block by height: %v", err)
		}

		h, err := buildBlockHalByHeight(hd, height)
		if err != nil {
			return nil, err
		}
		hal = h
	} else if s, found := vars["hash"]; found {
		h, err := parseHashFromPath(s)
		if err != nil {
			return nil, util.NewIDError("Bad request").Errorf("Invalid hash for block by hash: %v", err)
		}

		i, err := buildBlockHalByHash(hd, h)
		if err != nil {
			return nil, err
		}
		hal = i
	}

	return hd.enc.Marshal(hal)
}

func buildBlockHalByHeight(hd *Handlers, height base.Height) (Hal, error) {
	h, err := hd.CombineURL(HandlerPathBlockByHeight, "height", height.String())
	if err != nil {
		return nil, err
	}

	var hal Hal
	hal = NewBaseHal(nil, NewHalLink(h, nil))

	h, err = hd.CombineURL(HandlerPathBlockByHeight, "height", (height + 1).String())
	if err != nil {
		return nil, err
	}
	hal = hal.AddLink("next", NewHalLink(h, nil))

	if height > base.GenesisHeight {
		h, err = hd.CombineURL(HandlerPathBlockByHeight, "height", (height - 1).String())
		if err != nil {
			return nil, err
		}
		hal = hal.AddLink("prev", NewHalLink(h, nil))
	}

	h, err = hd.CombineURL(HandlerPathBlockByHeight, "height", height.String())
	if err != nil {
		return nil, err
	}
	hal = hal.AddLink("current", NewHalLink(h, nil))

	h, err = hd.CombineURL(HandlerPathManifestByHeight, "height", height.String())
	if err != nil {
		return nil, err
	}
	hal = hal.AddLink("current-manifest", NewHalLink(h, nil))

	for k := range halBlockTemplate {
		hal = hal.AddLink(k, halBlockTemplate[k])
	}

	return hal, nil
}

func buildBlockHalByHash(hd *Handlers, h util.Hash) (Hal, error) {
	i, err := hd.CombineURL(HandlerPathBlockByHash, "hash", h.String())
	if err != nil {
		return nil, err
	}

	var hal Hal
	hal = NewBaseHal(nil, NewHalLink(i, nil))

	i, err = hd.CombineURL(HandlerPathManifestByHash, "hash", h.String())
	if err != nil {
		return nil, err
	}
	hal = hal.AddLink("manifest", NewHalLink(i, nil))

	for k := range halBlockTemplate {
		hal = hal.AddLink(k, halBlockTemplate[k])
	}

	return hal, nil
}
