//go:build dev
// +build dev

package digest

import (
	"net/http"
	"runtime"
	"strings"
	"time"
)

func (hd *Handlers) handleResource(w http.ResponseWriter, r *http.Request) {
	memUnit := ParseStringQuery(r.URL.Query().Get("unit"))
	keys := ParseCSVStringQuery(strings.ToLower(r.URL.Query().Get("keys")))
	cacheKey := CacheKeyPath(r)
	if err := LoadFromCache(hd.cache, cacheKey, w); err == nil {
		return
	}

	if v, err, shared := hd.rg.Do(cacheKey, func() (interface{}, error) {
		return hd.handleResourceInGroup(memUnit, keys)
	}); err != nil {
		HTTP2HandleError(w, err)
	} else {
		HTTP2WriteHalBytes(hd.enc, w, v.([]byte), http.StatusOK)
		if !shared {
			HTTP2WriteCache(w, cacheKey, hd.expireShortLived)
		}
	}
}

func (hd *Handlers) handleResourceInGroup(unit string, keys []string) (interface{}, error) {
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

	var m struct {
		MemInfo map[string]MemoryMetric `json:"mem"`
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

	m.MemInfo = map[string]MemoryMetric{
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
		memInfo := make(map[string]MemoryMetric)
		for _, key := range keys {
			k, found := MemInfoKeys[key]
			if found {
				memInfo[k] = m.MemInfo[k]
			}
		}

		m.MemInfo = memInfo
	}

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

type MemoryMetric struct {
	Value       interface{} `json:"value"`
	Unit        string      `json:"unit"`
	Description string      `json:"description"`
}
