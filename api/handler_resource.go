//go:build dev
// +build dev

package api

import (
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type resourceMetrics struct {
	memInfo map[string]MemoryMetric
	raw     runtime.MemStats
}

func HandleResource(hd *Handlers, w http.ResponseWriter, r *http.Request) {
	memUnit := ParseStringQuery(r.URL.Query().Get("unit"))
	keys := ParseCSVStringQuery(strings.ToLower(r.URL.Query().Get("keys")))
	cacheKey := CacheKeyPath(r)
	if err := LoadFromCache(hd.cache, cacheKey, w); err == nil {
		return
	}

	if v, err, shared := hd.rg.Do(cacheKey, func() (interface{}, error) {
		return handleResourceInGroup(hd, memUnit, keys)
	}); err != nil {
		HTTP2HandleError(w, err)
	} else {
		HTTP2WriteHalBytes(hd.enc, w, v.([]byte), http.StatusOK)
		if !shared {
			HTTP2WriteCache(w, cacheKey, hd.expireShortLived)
		}
	}
}

func handleResourceInGroup(hd *Handlers, unit string, keys []string) (interface{}, error) {
	rm, err := collectResourceMetrics(hd, unit, keys)
	if err != nil {
		return nil, err
	}

	var payload struct {
		MemInfo map[string]MemoryMetric `json:"mem"`
	}

	payload.MemInfo = rm.memInfo

	hal, err := buildResourceHal(hd, payload)
	if err != nil {
		return nil, err
	}
	return hd.enc.Marshal(hal)
}

func collectResourceMetrics(hd *Handlers, unit string, keys []string) (*resourceMetrics, error) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	const B = 1
	const KB = 1024
	const MB = 1024 * 1024
	const GB = 1024 * 1024 * 1024
	var Unit float64
	var UnitStr string
	switch unit {
	case "KB", "kb":
		Unit = KB
		UnitStr = "KByte"
	case "MB", "mb":
		Unit = MB
		UnitStr = "MByte"
	case "GB", "gb":
		Unit = GB
		UnitStr = "GByte"
	default:
		Unit = B
		UnitStr = "Byte"
	}

	convert := func(b uint64) float64 {
		return float64(b) / Unit
	}

	MemInfoKeys := map[string]string{
		"alloc":         "Alloc",
		"totalalloc":    "TotalAlloc",
		"sys":           "Sys",
		"heapalloc":     "HeapAlloc",
		"heapsys":       "HeapSys",
		"heapidle":      "HeapIdle",
		"heapinuse":     "HeapInuse",
		"heapreleased":  "HeapReleased",
		"stackinuse":    "StackInuse",
		"stacksys":      "StackSys",
		"nextgc":        "NextGC",
		"heapobjects":   "HeapObjects",
		"mspaninuse":    "MSpanInuse",
		"mspansys":      "MSpanSys",
		"mcacheinuse":   "MCacheInuse",
		"mcachesys":     "MCacheSys",
		"buckhashsys":   "BuckHashSys",
		"gcsys":         "GCSys",
		"othersys":      "OtherSys",
		"lastgc":        "LastGC",
		"pausetotalns":  "PauseTotalNs",
		"numgc":         "NumGC",
		"numforcedgc":   "NumForcedGC",
		"gccpufraction": "GCCPUFraction",
		"enablegc":      "EnableGC",
		"debuggc":       "DebugGC",
	}

	memInfo := map[string]MemoryMetric{
		"Alloc": {
			Value:       convert(mem.Alloc),
			Unit:        UnitStr,
			Description: "현재 할당된 힙 메모리",
		},
		"TotalAlloc": {
			Value:       convert(mem.TotalAlloc),
			Unit:        UnitStr,
			Description: "프로그램 전체 실행 중 누적 할당된 힙 메모리",
		},
		"Sys": {
			Value:       convert(mem.Sys),
			Unit:        UnitStr,
			Description: "Go 런타임이 OS로부터 확보한 전체 메모리",
		},
		"HeapAlloc": {
			Value:       convert(mem.HeapAlloc),
			Unit:        UnitStr,
			Description: "현재 할당된 힙 메모리 (Alloc과 동일)",
		},
		"HeapSys": {
			Value:       convert(mem.HeapSys),
			Unit:        UnitStr,
			Description: "힙 용도로 확보한 전체 메모리",
		},
		"HeapIdle": {
			Value:       convert(mem.HeapIdle),
			Unit:        UnitStr,
			Description: "사용되지 않는 힙 메모리",
		},
		"HeapInuse": {
			Value:       convert(mem.HeapInuse),
			Unit:        UnitStr,
			Description: "현재 사용 중인 힙 메모리",
		},
		"HeapReleased": {
			Value:       convert(mem.HeapReleased),
			Unit:        UnitStr,
			Description: "OS에 반환된 힙 메모리",
		},
		"StackInuse": {
			Value:       convert(mem.StackInuse),
			Unit:        UnitStr,
			Description: "고루틴 스택에 사용된 메모리",
		},
		"StackSys": {
			Value:       convert(mem.StackSys),
			Unit:        UnitStr,
			Description: "스택 용도로 확보한 메모리",
		},
		"NextGC": {
			Value:       convert(mem.NextGC),
			Unit:        UnitStr,
			Description: "다음 GC 트리거 메모리 임계값",
		},
		"HeapObjects": {
			Value:       mem.HeapObjects,
			Unit:        "count",
			Description: "현재 살아 있는 힙 객체 수",
		},
		"MSpanInuse": {
			Value:       convert(mem.MSpanInuse),
			Unit:        UnitStr,
			Description: "런타임이 현재 사용 중인 mspan 메모리",
		},
		"MSpanSys": {
			Value:       convert(mem.MSpanSys),
			Unit:        UnitStr,
			Description: "mspan 용도로 확보된 전체 메모리",
		},
		"MCacheInuse": {
			Value:       convert(mem.MCacheInuse),
			Unit:        UnitStr,
			Description: "사용 중인 mcache 구조체 메모리",
		},
		"MCacheSys": {
			Value:       convert(mem.MCacheSys),
			Unit:        UnitStr,
			Description: "mcache 용도로 확보된 전체 메모리",
		},
		"BuckHashSys": {
			Value:       convert(mem.BuckHashSys),
			Unit:        UnitStr,
			Description: "버킷 해시 테이블 용 메모리 (profile용)",
		},
		"GCSys": {
			Value:       convert(mem.GCSys),
			Unit:        UnitStr,
			Description: "GC 메타데이터에 사용된 메모리",
		},
		"OtherSys": {
			Value:       convert(mem.OtherSys),
			Unit:        UnitStr,
			Description: "기타 런타임 시스템 메모리",
		},
		"LastGC": {
			Value:       time.Unix(0, int64(mem.LastGC)).Format(time.RFC3339Nano),
			Unit:        "timestamp",
			Description: "마지막 GC가 끝난 시간",
		},
		"PauseTotalNs": {
			Value:       mem.PauseTotalNs,
			Unit:        "ns",
			Description: "총 GC 중단 시간 (누적)",
		},
		"NumGC": {
			Value:       mem.NumGC,
			Unit:        "count",
			Description: "총 GC 수행 횟수",
		},
		"NumForcedGC": {
			Value:       mem.NumForcedGC,
			Unit:        "count",
			Description: "프로그래밍적으로 호출된 강제 GC 횟수",
		},
		"GCCPUFraction": {
			Value:       mem.GCCPUFraction,
			Unit:        "fraction",
			Description: "프로그램이 GC에 소비한 CPU 시간 비율",
		},
		"EnableGC": {
			Value:       mem.EnableGC,
			Unit:        "bool",
			Description: "GC 사용 여부",
		},
		"DebugGC": {
			Value:       mem.DebugGC,
			Unit:        "bool",
			Description: "디버그용 GC 설정 여부 (현재 미사용)",
		},
	}

	switch {
	case len(keys) == 1 && keys[0] == "":
	case len(keys) < 1:
	default:
		selected := make(map[string]MemoryMetric)
		for _, key := range keys {
			k, found := MemInfoKeys[key]
			if found {
				if metric, ok := memInfo[k]; ok {
					selected[k] = metric
				}
			}
		}

		memInfo = selected
	}

	return &resourceMetrics{
		memInfo: memInfo,
		raw:     mem,
	}, nil
}

func HandleResourceProm(hd *Handlers, w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	memUnit := ParseStringQuery(r.URL.Query().Get("unit"))
	keys := ParseCSVStringQuery(strings.ToLower(r.URL.Query().Get("keys")))

	rm, err := collectResourceMetrics(hd, memUnit, keys)
	if err != nil {
		HTTP2HandleError(w, err)

		return
	}

	var b strings.Builder
	writePromResource(&b, rm)

	w.Header().Set("Content-Type", PrometheusTextMimetype)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(b.String()))
}

func buildResourceHal(hd *Handlers, resource interface{}) (Hal, error) {
	hal := NewBaseHal(resource, NewHalLink(HandlerPathResource, nil))

	return hal, nil
}

type resourcePromMetric struct {
	key   string
	name  string
	value func(mem *runtime.MemStats) (string, bool)
}

var resourcePromMetrics = []resourcePromMetric{
	{
		key:  "Alloc",
		name: "mitum_resource_memory_alloc_bytes",
		value: func(mem *runtime.MemStats) (string, bool) {
			return strconv.FormatUint(mem.Alloc, 10), true
		},
	},
	{
		key:  "TotalAlloc",
		name: "mitum_resource_memory_total_alloc_bytes",
		value: func(mem *runtime.MemStats) (string, bool) {
			return strconv.FormatUint(mem.TotalAlloc, 10), true
		},
	},
	{
		key:  "Sys",
		name: "mitum_resource_memory_sys_bytes",
		value: func(mem *runtime.MemStats) (string, bool) {
			return strconv.FormatUint(mem.Sys, 10), true
		},
	},
	{
		key:  "HeapAlloc",
		name: "mitum_resource_memory_heap_alloc_bytes",
		value: func(mem *runtime.MemStats) (string, bool) {
			return strconv.FormatUint(mem.HeapAlloc, 10), true
		},
	},
	{
		key:  "HeapSys",
		name: "mitum_resource_memory_heap_sys_bytes",
		value: func(mem *runtime.MemStats) (string, bool) {
			return strconv.FormatUint(mem.HeapSys, 10), true
		},
	},
	{
		key:  "HeapIdle",
		name: "mitum_resource_memory_heap_idle_bytes",
		value: func(mem *runtime.MemStats) (string, bool) {
			return strconv.FormatUint(mem.HeapIdle, 10), true
		},
	},
	{
		key:  "HeapInuse",
		name: "mitum_resource_memory_heap_inuse_bytes",
		value: func(mem *runtime.MemStats) (string, bool) {
			return strconv.FormatUint(mem.HeapInuse, 10), true
		},
	},
	{
		key:  "HeapReleased",
		name: "mitum_resource_memory_heap_released_bytes",
		value: func(mem *runtime.MemStats) (string, bool) {
			return strconv.FormatUint(mem.HeapReleased, 10), true
		},
	},
	{
		key:  "StackInuse",
		name: "mitum_resource_memory_stack_inuse_bytes",
		value: func(mem *runtime.MemStats) (string, bool) {
			return strconv.FormatUint(mem.StackInuse, 10), true
		},
	},
	{
		key:  "StackSys",
		name: "mitum_resource_memory_stack_sys_bytes",
		value: func(mem *runtime.MemStats) (string, bool) {
			return strconv.FormatUint(mem.StackSys, 10), true
		},
	},
	{
		key:  "NextGC",
		name: "mitum_resource_memory_next_gc_bytes",
		value: func(mem *runtime.MemStats) (string, bool) {
			return strconv.FormatUint(mem.NextGC, 10), true
		},
	},
	{
		key:  "HeapObjects",
		name: "mitum_resource_memory_heap_objects",
		value: func(mem *runtime.MemStats) (string, bool) {
			return strconv.FormatUint(mem.HeapObjects, 10), true
		},
	},
	{
		key:  "MSpanInuse",
		name: "mitum_resource_memory_mspan_inuse_bytes",
		value: func(mem *runtime.MemStats) (string, bool) {
			return strconv.FormatUint(mem.MSpanInuse, 10), true
		},
	},
	{
		key:  "MSpanSys",
		name: "mitum_resource_memory_mspan_sys_bytes",
		value: func(mem *runtime.MemStats) (string, bool) {
			return strconv.FormatUint(mem.MSpanSys, 10), true
		},
	},
	{
		key:  "MCacheInuse",
		name: "mitum_resource_memory_mcache_inuse_bytes",
		value: func(mem *runtime.MemStats) (string, bool) {
			return strconv.FormatUint(mem.MCacheInuse, 10), true
		},
	},
	{
		key:  "MCacheSys",
		name: "mitum_resource_memory_mcache_sys_bytes",
		value: func(mem *runtime.MemStats) (string, bool) {
			return strconv.FormatUint(mem.MCacheSys, 10), true
		},
	},
	{
		key:  "BuckHashSys",
		name: "mitum_resource_memory_buck_hash_sys_bytes",
		value: func(mem *runtime.MemStats) (string, bool) {
			return strconv.FormatUint(mem.BuckHashSys, 10), true
		},
	},
	{
		key:  "GCSys",
		name: "mitum_resource_memory_gc_sys_bytes",
		value: func(mem *runtime.MemStats) (string, bool) {
			return strconv.FormatUint(mem.GCSys, 10), true
		},
	},
	{
		key:  "OtherSys",
		name: "mitum_resource_memory_other_sys_bytes",
		value: func(mem *runtime.MemStats) (string, bool) {
			return strconv.FormatUint(mem.OtherSys, 10), true
		},
	},
	{
		key:  "LastGC",
		name: "mitum_resource_memory_last_gc_timestamp_seconds",
		value: func(mem *runtime.MemStats) (string, bool) {
			return strconv.FormatFloat(float64(mem.LastGC)/1e9, 'f', -1, 64), true
		},
	},
	{
		key:  "PauseTotalNs",
		name: "mitum_resource_memory_pause_total_seconds",
		value: func(mem *runtime.MemStats) (string, bool) {
			return strconv.FormatFloat(float64(mem.PauseTotalNs)/1e9, 'f', -1, 64), true
		},
	},
	{
		key:  "NumGC",
		name: "mitum_resource_memory_num_gc_total",
		value: func(mem *runtime.MemStats) (string, bool) {
			return strconv.FormatUint(uint64(mem.NumGC), 10), true
		},
	},
	{
		key:  "NumForcedGC",
		name: "mitum_resource_memory_num_forced_gc_total",
		value: func(mem *runtime.MemStats) (string, bool) {
			return strconv.FormatUint(uint64(mem.NumForcedGC), 10), true
		},
	},
	{
		key:  "GCCPUFraction",
		name: "mitum_resource_memory_gc_cpu_fraction",
		value: func(mem *runtime.MemStats) (string, bool) {
			return strconv.FormatFloat(mem.GCCPUFraction, 'f', -1, 64), true
		},
	},
	{
		key:  "EnableGC",
		name: "mitum_resource_memory_enable_gc",
		value: func(mem *runtime.MemStats) (string, bool) {
			return boolToGauge(mem.EnableGC), true
		},
	},
	{
		key:  "DebugGC",
		name: "mitum_resource_memory_debug_gc",
		value: func(mem *runtime.MemStats) (string, bool) {
			return boolToGauge(mem.DebugGC), true
		},
	},
}

func writePromResource(b *strings.Builder, rm *resourceMetrics) {
	headersWritten := map[string]bool{}

	if rm == nil || len(rm.memInfo) == 0 {
		b.WriteString("# No resource metrics available\n")

		return
	}

	for i := range resourcePromMetrics {
		def := resourcePromMetrics[i]
		if _, ok := rm.memInfo[def.key]; !ok {
			continue
		}

		if value, ok := def.value(&rm.raw); ok {
			writePromSample(b, def.name, nil, value, headersWritten)
		}
	}
}

func boolToGauge(v bool) string {
	if v {
		return "1"
	}

	return "0"
}

type MemoryMetric struct {
	Value       interface{} `json:"value"`
	Unit        string      `json:"unit"`
	Description string      `json:"description"`
}
