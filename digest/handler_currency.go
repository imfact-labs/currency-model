package digest

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

func HandleCurrencies(hd *Handlers, w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	cachekey := CacheKeyPath(r)
	if err := LoadFromCache(hd.cache, cachekey, w); err == nil {
		return
	}

	if v, err, shared := hd.rg.Do(cachekey, func() (interface{}, error) {
		return handleCurrenciesInGroup(hd)
	}); err != nil {
		HTTP2HandleError(w, err)
	} else {
		HTTP2WriteHalBytes(hd.enc, w, v.([]byte), http.StatusOK)

		if !shared {
			HTTP2WriteCache(w, cachekey, hd.expireShortLived)
		}
	}
}

func handleCurrenciesInGroup(hd *Handlers) ([]byte, error) {
	var hal Hal = NewBaseHal(nil, NewHalLink(HandlerPathCurrencies, nil))
	hal = hal.AddLink("currency:{currency_id}", NewHalLink(HandlerPathCurrency, nil).SetTemplated())

	cids, err := hd.database.currencies()
	if err != nil {
		return nil, err
	}
	for i := range cids {
		h, err := hd.CombineURL(HandlerPathCurrency, "currency_id", cids[i])
		if err != nil {
			return nil, err
		}
		hal = hal.AddLink(fmt.Sprintf("currency:%s", cids[i]), NewHalLink(h, nil))
	}

	return hd.enc.Marshal(hal)
}

func HandleCurrency(hd *Handlers, w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	cachekey := CacheKeyPath(r)
	if err := LoadFromCache(hd.cache, cachekey, w); err == nil {
		return
	}

	var cid string
	s, found := mux.Vars(r)["currency_id"]
	if !found {
		HTTP2ProblemWithError(w, errors.Errorf("Empty currency id"), http.StatusBadRequest)

		return
	}

	s = strings.TrimSpace(s)
	if len(s) < 1 {
		HTTP2ProblemWithError(w, errors.Errorf("Empty currency id"), http.StatusBadRequest)

		return
	}
	cid = s

	if v, err, shared := hd.rg.Do(cachekey, func() (interface{}, error) {
		return handleCurrencyInGroup(hd, cid)
	}); err != nil {
		HTTP2HandleError(w, err)
	} else {
		HTTP2WriteHalBytes(hd.enc, w, v.([]byte), http.StatusOK)

		if !shared {
			HTTP2WriteCache(w, cachekey, hd.expireShortLived)
		}
	}
}

func handleCurrencyInGroup(hd *Handlers, cid string) ([]byte, error) {
	var de types.CurrencyDesign
	var st base.State

	de, st, err := hd.database.currency(cid)
	if err != nil {
		return nil, err
	}

	i, err := buildCurrency(hd, de, st)
	if err != nil {
		return nil, err
	}
	return hd.enc.Marshal(i)
}

func buildCurrency(hd *Handlers, de types.CurrencyDesign, st base.State) (Hal, error) {
	h, err := hd.CombineURL(HandlerPathCurrency, "currency_id", de.Currency().String())
	if err != nil {
		return nil, err
	}

	var hal Hal
	hal = NewBaseHal(de, NewHalLink(h, nil))

	hal = hal.AddLink("currency:{currency_id}", NewHalLink(HandlerPathCurrency, nil).SetTemplated())

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
