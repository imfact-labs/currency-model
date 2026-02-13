package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/imfact-labs/imfact-currency/digest"
	"github.com/imfact-labs/imfact-currency/utils"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/valuehash"
	"github.com/gorilla/mux"
	"github.com/json-iterator/go"
	"github.com/pkg/errors"
)

func buildOffset(height base.Height, index uint64) string {
	return fmt.Sprintf("%d,%d", height, index)
}

func buildOffsetByString(height base.Height, s string) string {
	return fmt.Sprintf("%d,%s", height, s)
}

func parseOffset(s string) (base.Height, uint64, error) {
	if n := strings.SplitN(s, ",", 2); n == nil {
		return base.NilHeight, 0, errors.Errorf("Invalid offset string, %q", s)
	} else if len(n) < 2 {
		return base.NilHeight, 0, errors.Errorf("Invalid offset, %q", s)
	} else if h, err := base.ParseHeightString(n[0]); err != nil {
		return base.NilHeight, 0, errors.Wrap(err, "invalid height of offset")
	} else if u, err := strconv.ParseUint(n[1], 10, 64); err != nil {
		return base.NilHeight, 0, errors.Wrap(err, "invalid index of offset")
	} else {
		return h, u, nil
	}
}

func parseOffsetByString(s string) (base.Height, string, error) {
	var a, b string
	switch n := strings.SplitN(s, ",", 2); {
	case n == nil:
		return base.NilHeight, "", errors.Errorf("Invalid offset string, %q", s)
	case len(n) < 2:
		return base.NilHeight, "", errors.Errorf("Invalid offset, %q", s)
	default:
		a = n[0]
		b = n[1]
	}

	h, err := base.ParseHeightString(a)
	if err != nil {
		return base.NilHeight, "", errors.Wrap(err, "invalid height of offset")
	}

	return h, b, nil
}

func parseHeightFromPath(s string) (base.Height, error) {
	s = strings.TrimSpace(s)

	if len(s) < 1 {
		return base.NilHeight, errors.Errorf("Empty height")
	} else if len(s) > 1 && strings.HasPrefix(s, "0") {
		return base.NilHeight, errors.Errorf("Invalid height, %v", s)
	}

	return base.ParseHeightString(s)
}

func parseHashFromPath(s string) (util.Hash, error) {
	s = strings.TrimSpace(s)
	if len(s) < 1 {
		return nil, errors.Errorf("Empty hash")
	}

	h := valuehash.NewBytesFromString(s)

	err := h.IsValid(nil)
	if err != nil {
		return nil, err
	}

	return h, nil
}

func ParseLimitQuery(s string) int64 {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return int64(-1)
	}
	return n
}

func ParseStringQuery(s string) string {
	return strings.TrimSpace(s)
}

func ParseCSVStringQuery(s string) []string {
	return strings.Split(strings.TrimSpace(s), ",")
}

func ParseRequest(_ http.ResponseWriter, r *http.Request, v string) (string, error, int) {
	s, found := mux.Vars(r)[v]
	if !found {
		return "", errors.Errorf("empty %s", v), http.StatusNotFound
	}

	s = strings.TrimSpace(s)
	if len(s) < 1 {
		return "", errors.Errorf("empty %s", v), http.StatusBadRequest
	}
	return s, nil, http.StatusOK
}

func StringOffsetQuery(offset string) string {
	return fmt.Sprintf("offset=%s", offset)
}

func stringCurrencyQuery(currencyId string) string {
	return fmt.Sprintf("currency=%s", currencyId)
}

func stringHashesQuery(hashes string) string {
	return fmt.Sprintf("hashes=%s", hashes)
}

func ParseBoolQuery(s string) bool {
	return s == "1" || s == "true"
}

func StringBoolQuery(key string, v bool) string { // nolint:unparam
	if v {
		return fmt.Sprintf("%s=1", key)
	}

	return ""
}

func AddQueryValue(b, s string) string {
	if len(s) < 1 {
		return b
	}

	if !strings.Contains(b, "?") {
		return b + "?" + s
	}

	return b + "&" + s
}

func HTTP2Stream(enc encoder.Encoder, w http.ResponseWriter, bufsize int, status int) (*jsoniter.Stream, func()) {
	w.Header().Set(HTTP2EncoderHintHeader, enc.Hint().String())
	w.Header().Set("Content-Type", HALMimetype)

	if status != http.StatusOK {
		w.WriteHeader(status)
	}

	stream := jsoniter.NewStream(HALJSONConfigDefault, w, bufsize)
	return stream, func() {
		_ = stream.Flush()
	}
}

func HTTP2NotSupported(w http.ResponseWriter, err error) {
	if err == nil {
		err = util.NewIDError("not supported")
	}

	HTTP2ProblemWithError(w, err, http.StatusInternalServerError)
}

func HTTP2ProblemWithError(w http.ResponseWriter, err error, status int) {
	HTTP2WriteProblem(w, NewProblemFromError(err), status)
}

func HTTP2WriteProblem(w http.ResponseWriter, pr Problem, status int) {
	if status == 0 {
		status = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", ProblemMimetype)
	w.Header().Set("X-Content-Type-Options", "nosniff")

	var output []byte
	if b, err := utils.JSON.Marshal(pr.title); err != nil {
		output = digest.UnknownProblemJSON
	} else {
		output = b
	}

	w.WriteHeader(status)
	_, _ = w.Write(output)
}

func HTTP2WriteHal(enc encoder.Encoder, w http.ResponseWriter, hal Hal, status int) { // nolint:unparam
	stream, flush := HTTP2Stream(enc, w, 1, status)
	defer flush()

	stream.WriteVal(hal)
}

func HTTP2WriteHalBytes(enc encoder.Encoder, w http.ResponseWriter, b []byte, status int) { // nolint:unparam
	w.Header().Set(HTTP2EncoderHintHeader, enc.Hint().String())
	w.Header().Set("Content-Type", HALMimetype)

	if status != http.StatusOK {
		w.WriteHeader(status)
	}

	_, _ = w.Write(b)
}

func HTTP2WriteBytes(w http.ResponseWriter, b []byte, contentType string, status int) {
	w.Header().Set("Content-Type", contentType)

	if status != http.StatusOK {
		w.WriteHeader(status)
	}

	_, _ = w.Write(b)
}

func HTTP2WriteCache(w http.ResponseWriter, key string, expire time.Duration) {
	if expire < 1 {
		return
	}

	if cw, ok := w.(*CacheResponseWriter); ok {
		_ = cw.SetKey(key).SetExpire(expire)
	}
}

func HTTP2HandleError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, util.ErrNotFound):
		status = http.StatusBadRequest
	case errors.Is(err, digest.ErrBadRequest):
		status = http.StatusBadRequest
	case errors.Is(err, util.NewIDError("not supported")):
		status = http.StatusInternalServerError
	}

	HTTP2ProblemWithError(w, err, status)
}
