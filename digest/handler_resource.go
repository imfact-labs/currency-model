//go:build dev
// +build dev

package digest

import (
	"net/http"
	"runtime"
)

func (hd *Handlers) handleResource(w http.ResponseWriter, r *http.Request) {
	cacheKey := CacheKeyPath(r)
	if err := LoadFromCache(hd.cache, cacheKey, w); err == nil {
		return
	}

	if v, err, shared := hd.rg.Do(cacheKey, func() (interface{}, error) {
		return hd.handleResourceInGroup()
	}); err != nil {
		HTTP2HandleError(w, err)
	} else {
		HTTP2WriteHalBytes(hd.enc, w, v.([]byte), http.StatusOK)
		if !shared {
			HTTP2WriteCache(w, cacheKey, hd.expireShortLived)
		}
	}
}

func (hd *Handlers) handleResourceInGroup() (interface{}, error) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	var m struct {
		MemInfo runtime.MemStats `json:"mem"`
	}

	m.MemInfo = mem

	hal, err := hd.buildResourceHal(m)
	if err != nil {
		return nil, err
	}
	return hd.enc.Marshal(hal)

}

func (hd *Handlers) buildResourceHal(resource interface{}) (Hal, error) {
	hal := NewBaseHal(resource, NewHalLink(HandlerPathResource, nil))

	return hal, nil
}
