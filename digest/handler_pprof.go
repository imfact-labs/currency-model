//go:build dev
// +build dev

package digest

import (
	"bytes"
	"fmt"
	"github.com/google/pprof/profile"
	"io"
	"net/http"
	"regexp"
	"runtime/pprof"
	"sort"
	"time"
)

var defaultCount int = 10

func (hd *Handlers) handlePProfProfile(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	duration := ParseLimitQuery(r.URL.Query().Get("duration"))
	count := ParseLimitQuery(r.URL.Query().Get("count"))
	cacheKey := CacheKeyPath(r)
	if err := LoadFromCache(hd.cache, cacheKey, w); err == nil {
		return
	}

	if v, err, shared := hd.rg.Do(cacheKey, func() (interface{}, error) {
		return hd.handlePProfProfileInGroup(duration, count)
	}); err != nil {
		HTTP2HandleError(w, err)
	} else {
		HTTP2WriteBytes(w, v.([]byte), PlainTextMimetype, http.StatusOK)
		if !shared {
			HTTP2WriteCache(w, cacheKey, hd.expireShortLived)
		}
	}
}

func (hd *Handlers) handlePProfProfileInGroup(du, cnt int64) (interface{}, error) {
	var duration int
	count := defaultCount
	if du < 0 {
		duration = 10
	} else {
		duration = int(du)
	}
	if cnt > 0 {
		count = int(cnt)
	}

	pr, pw := io.Pipe()

	if err := pprof.StartCPUProfile(pw); err != nil {
		return nil, err
	}
	go func(dur int) {
		defer pw.Close()
		time.Sleep(time.Duration(dur) * time.Second)
		pprof.StopCPUProfile()
	}(duration)

	prof, err := profile.Parse(pr)
	if err != nil {
		return nil, err
	}

	lines, err := parsePProfCPUTop(prof, "cpu", count)
	if err != nil {
		return nil, err
	}

	var out bytes.Buffer
	for _, line := range lines {
		out.WriteString(line + "\n")
	}

	return out.Bytes(), nil
}

type profileStat struct {
	Name    string
	Flat    int64
	Cum     int64
	FlatPct float64
	CumPct  float64
}

func parsePProfCPUTop(prof *profile.Profile, sampleType string, topN int) ([]string, error) {
	sampleIndex := -1
	unit := ""
	for i, s := range prof.SampleType {
		if s.Type == sampleType {
			sampleIndex = i
			unit = s.Unit
			break
		}
	}
	if sampleIndex < 0 {
		return nil, fmt.Errorf("sample type %q not found", sampleType)
	}

	total := int64(0)
	flatCount := map[string]int64{}
	cumCount := map[string]int64{}

	for _, sample := range prof.Sample {
		v := sample.Value[sampleIndex]
		total += v

		seen := map[string]bool{}
		for i := len(sample.Location) - 1; i >= 0; i-- {
			loc := sample.Location[i]
			for _, line := range loc.Line {
				name := line.Function.Name
				if !seen[name] {
					cumCount[name] += v
					seen[name] = true
				}
			}
		}

		if len(sample.Location) > 0 {
			topLoc := sample.Location[0]
			for _, line := range topLoc.Line {
				name := line.Function.Name
				flatCount[name] += v
			}
		}
	}

	var stats []profileStat
	for name, flat := range flatCount {
		cum := cumCount[name]
		stats = append(stats, profileStat{
			Name:    name,
			Flat:    flat,
			Cum:     cum,
			FlatPct: float64(flat) * 100 / float64(total),
			CumPct:  float64(cum) * 100 / float64(total),
		})
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Flat > stats[j].Flat
	})

	var out []string
	out = append(out,
		fmt.Sprintf("Showing nodes accounting for %s, 100%% of %s total",
			formatProfileValue(total, unit),
			formatProfileValue(total, unit),
		),
	)

	if topN > len(stats) {
		topN = len(stats)
	}

	out = append(out, fmt.Sprintf("Showing top %d nodes out of %d", topN, len(stats)))
	out = append(out, fmt.Sprintf("%8s %7s %6s %8s %6s  %s", "flat", "flat%", "sum%", "cum", "cum%", "function"))

	sumPct := 0.0
	for i := 0; i < topN && i < len(stats); i++ {
		s := stats[i]
		sumPct += s.FlatPct
		out = append(out, fmt.Sprintf(
			"%8s %6.2f%% %6.2f%% %8s %6.2f%%  %s",
			formatProfileValue(s.Flat, unit),
			s.FlatPct,
			sumPct,
			formatProfileValue(s.Cum, unit),
			s.CumPct,
			s.Name,
		))
	}

	return out, nil
}

func formatProfileValue(v int64, unit string) string {
	switch unit {
	case "nanoseconds", "ns":
		return fmt.Sprintf("%dms", v/1e6)
	default:
		return fmt.Sprintf("%d", v)
	}
}

type funcStat struct {
	Name    string
	Flat    int64
	Cum     int64
	FlatPct float64
	CumPct  float64
}

func (hd *Handlers) handlePProfHeap(w http.ResponseWriter, r *http.Request) {
	count := ParseLimitQuery(r.URL.Query().Get("count"))
	cacheKey := CacheKeyPath(r)
	if err := LoadFromCache(hd.cache, cacheKey, w); err == nil {
		return
	}

	if v, err, shared := hd.rg.Do(cacheKey, func() (interface{}, error) {
		return hd.handlePProfHeapInGroup(count)
	}); err != nil {
		HTTP2HandleError(w, err)
	} else {
		HTTP2WriteBytes(w, v.([]byte), PlainTextMimetype, http.StatusOK)
		if !shared {
			HTTP2WriteCache(w, cacheKey, hd.expireShortLived)
		}
	}
}

func (hd *Handlers) handlePProfHeapInGroup(cnt int64) (interface{}, error) {
	count := defaultCount
	if cnt > 0 {
		count = int(cnt)
	}

	var buf bytes.Buffer
	if err := pprof.Lookup("heap").WriteTo(&buf, 0); err != nil {
		return nil, err
	}

	prof, err := profile.Parse(&buf)
	if err != nil {
		return nil, err
	}

	lines, err := parsePProfHeapAllocsTop(prof, "inuse_space", count)
	if err != nil {
		return nil, err
	}

	var out bytes.Buffer
	for _, line := range lines {
		out.WriteString(line + "\n")
	}

	return out.Bytes(), nil
}

func parsePProfHeapAllocsTop(prof *profile.Profile, sampleType string, topN int) ([]string, error) {
	sampleIndex := -1
	unit := ""
	for i, s := range prof.SampleType {
		if s.Type == sampleType {
			sampleIndex = i
			unit = s.Unit
			break
		}
	}
	if sampleIndex == -1 {
		return nil, fmt.Errorf("sample type %q not found", sampleType)
	}

	total := int64(0)
	funcStats := map[string]*funcStat{}

	for _, sample := range prof.Sample {
		if len(sample.Value) <= sampleIndex {
			continue
		}
		v := sample.Value[sampleIndex]
		total += v
		seen := map[string]bool{}
		for _, loc := range sample.Location {
			for _, line := range loc.Line {
				if line.Function == nil {
					continue
				}
				name := line.Function.Name
				if _, ok := funcStats[name]; !ok {
					funcStats[name] = &funcStat{Name: name}
				}
				if !seen[name] {
					funcStats[name].Cum += v
					seen[name] = true
				}
			}
		}
		if len(sample.Location) > 0 && len(sample.Location[0].Line) > 0 {
			fn := sample.Location[0].Line[0].Function
			if fn != nil {
				funcStats[fn.Name].Flat += v
			}
		}
	}

	var stats []funcStat
	for _, s := range funcStats {
		s.FlatPct = float64(s.Flat) * 100 / float64(total)
		s.CumPct = float64(s.Cum) * 100 / float64(total)
		stats = append(stats, *s)
	}
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Flat > stats[j].Flat
	})

	dropThreshold := total / 100
	var kept, dropped []funcStat
	for _, s := range stats {
		if s.Cum <= dropThreshold {
			dropped = append(dropped, s)
		} else {
			kept = append(kept, s)
		}
	}

	accountingTotal := int64(0)
	for _, s := range kept {
		accountingTotal += s.Flat
	}
	accountingPct := float64(accountingTotal) * 100 / float64(total)

	var out []string
	out = append(out,
		fmt.Sprintf("Showing nodes accounting for %s, %.2f%% of %s total",
			formatBytes(accountingTotal, unit),
			accountingPct,
			formatBytes(total, unit)),
	)

	if len(dropped) > 0 {
		out = append(out, fmt.Sprintf("Dropped %d nodes (cum <= %s)", len(dropped), formatBytes(dropThreshold, unit)))
	}

	out = append(out, fmt.Sprintf("Showing top %d nodes out of %d", topN, len(kept)))
	out = append(out, fmt.Sprintf("%8s %7s %6s %8s %6s  %s", "flat", "flat%", "sum%", "cum", "cum%", "function"))

	sumPct := 0.0
	for i := 0; i < topN && i < len(kept); i++ {
		s := kept[i]
		sumPct += s.FlatPct
		out = append(out, fmt.Sprintf(
			"%8s %6.2f%% %6.2f%% %8s %6.2f%%  %s",
			formatBytes(s.Flat, unit),
			s.FlatPct,
			sumPct,
			formatBytes(s.Cum, unit),
			s.CumPct,
			s.Name,
		))
	}

	return out, nil
}

func formatBytes(v int64, unit string) string {
	if unit != "bytes" {
		return fmt.Sprintf("%d", v)
	}

	f := float64(v)
	switch {
	case f > 1<<30:
		return fmt.Sprintf("%.2fGB", f/(1<<30))
	case f > 1<<20:
		return fmt.Sprintf("%.2fMB", f/(1<<20))
	case f > 1<<10:
		return fmt.Sprintf("%.2fkB", f/(1<<10))
	default:
		return fmt.Sprintf("%dB", v)
	}
}

func (hd *Handlers) handlePProfAllocs(w http.ResponseWriter, r *http.Request) {
	count := ParseLimitQuery(r.URL.Query().Get("count"))
	cacheKey := CacheKeyPath(r)
	if err := LoadFromCache(hd.cache, cacheKey, w); err == nil {
		return
	}

	if v, err, shared := hd.rg.Do(cacheKey, func() (interface{}, error) {
		return hd.handlePProfAllocsInGroup(count)
	}); err != nil {
		HTTP2HandleError(w, err)
	} else {
		HTTP2WriteBytes(w, v.([]byte), PlainTextMimetype, http.StatusOK)
		if !shared {
			HTTP2WriteCache(w, cacheKey, hd.expireShortLived)
		}
	}
}

func (hd *Handlers) handlePProfAllocsInGroup(cnt int64) (interface{}, error) {
	count := defaultCount
	if cnt > 0 {
		count = int(cnt)
	}

	var buf bytes.Buffer
	if err := pprof.Lookup("allocs").WriteTo(&buf, 0); err != nil {
		return nil, err
	}

	prof, err := profile.Parse(&buf)
	if err != nil {
		return nil, err
	}

	lines, err := parsePProfHeapAllocsTop(prof, "alloc_space", count)
	if err != nil {
		return nil, err
	}

	var out bytes.Buffer
	for _, line := range lines {
		out.WriteString(line + "\n")
	}

	return out.Bytes(), nil
}

type goroutineStat struct {
	Name string
	Flat int64
	Cum  int64
}

func (hd *Handlers) handlePProfGoroutine(w http.ResponseWriter, r *http.Request) {
	count := ParseLimitQuery(r.URL.Query().Get("count"))
	cacheKey := CacheKeyPath(r)
	if err := LoadFromCache(hd.cache, cacheKey, w); err == nil {
		return
	}

	if v, err, shared := hd.rg.Do(cacheKey, func() (interface{}, error) {
		return hd.handlePProfGoroutineInGroup(count)
	}); err != nil {
		HTTP2HandleError(w, err)
	} else {
		HTTP2WriteBytes(w, v.([]byte), PlainTextMimetype, http.StatusOK)
		if !shared {
			HTTP2WriteCache(w, cacheKey, hd.expireShortLived)
		}
	}
}

func (hd *Handlers) handlePProfGoroutineInGroup(cnt int64) (interface{}, error) {
	count := defaultCount
	if cnt > 0 {
		count = int(cnt)
	}

	var buf bytes.Buffer
	if err := pprof.Lookup("goroutine").WriteTo(&buf, 0); err != nil {
		return nil, err
	}

	p, err := profile.Parse(&buf)
	if err != nil {
		return nil, err
	}

	lines := parsePProfGoroutineTop(p, count)

	var out bytes.Buffer
	for _, line := range lines {
		out.WriteString(line + "\n")
	}

	return out.Bytes(), nil
}

var ignoreRE = regexp.MustCompile(`(internal|selectgo|chanrecv|netpollblock)`)

func parsePProfGoroutineTop(p *profile.Profile, topN int) []string {
	flatCnt := map[string]int64{}
	cumCnt := map[string]int64{}
	var total int64

	for _, s := range p.Sample {
		count := s.Value[0]
		total += count

		seenInSample := map[string]bool{}

		for i, loc := range s.Location {
			for _, line := range loc.Line {
				fn := line.Function.Name

				if i == 0 {
					flatCnt[fn] += count
				}

				if !seenInSample[fn] {
					cumCnt[fn] += count
					seenInSample[fn] = true
				}
			}
		}
	}

	stats := make([]goroutineStat, 0, len(cumCnt))
	var sumFlat int64
	for fn, c := range cumCnt {
		if ignoreRE.MatchString(fn) {
			continue
		}

		flatValue := flatCnt[fn]
		stats = append(stats, goroutineStat{
			Name: fn,
			Flat: flatValue,
			Cum:  c,
		})
		sumFlat += flatValue
	}

	sort.Slice(stats, func(i, j int) bool {
		if stats[i].Flat != stats[j].Flat {
			return stats[i].Flat > stats[j].Flat
		}

		if stats[i].Cum != stats[j].Cum {
			return stats[i].Cum > stats[j].Cum
		}
		return stats[i].Name < stats[j].Name
	})

	if topN < 0 || topN > len(stats) {
		topN = len(stats)
	}

	out := []string{
		fmt.Sprintf("Showing nodes accounting for %d, %.2f%% of %d total",
			sumFlat, float64(sumFlat)*100/float64(total), total),
		fmt.Sprintf("Showing top %d nodes out of %d", topN, len(stats)),
		"      flat  flat%   sum%        cum   cum%  function",
	}

	var cumSum int64
	for i := 0; i < topN; i++ {
		s := stats[i]
		cumSum += s.Flat

		flatPct := float64(s.Flat) * 100 / float64(total)
		sumPct := float64(cumSum) * 100 / float64(total)
		cumPct := float64(s.Cum) * 100 / float64(total)

		out = append(out, fmt.Sprintf(
			"%8d %6.2f%% %6.2f%% %8d %6.2f%%  %s",
			s.Flat,
			flatPct,
			sumPct,
			s.Cum,
			cumPct,
			s.Name,
		))
	}

	return out
}

type blockStat struct {
	Name      string
	FlatDelay time.Duration
	CumDelay  time.Duration
	FlatCount int64
	CumCount  int64
}

func (hd *Handlers) handlePProfBlock(w http.ResponseWriter, r *http.Request) {
	count := ParseLimitQuery(r.URL.Query().Get("count"))
	cacheKey := CacheKeyPath(r)
	if err := LoadFromCache(hd.cache, cacheKey, w); err == nil {
		return
	}

	if v, err, shared := hd.rg.Do(cacheKey, func() (interface{}, error) {
		return hd.handlePProfBlockInGroup(count)
	}); err != nil {
		HTTP2HandleError(w, err)
	} else {
		HTTP2WriteBytes(w, v.([]byte), PlainTextMimetype, http.StatusOK)
		if !shared {
			HTTP2WriteCache(w, cacheKey, hd.expireShortLived)
		}
	}
}

func (hd *Handlers) handlePProfBlockInGroup(cnt int64) (interface{}, error) {
	count := defaultCount
	if cnt > 0 {
		count = int(cnt)
	}

	var buf bytes.Buffer
	if err := pprof.Lookup("block").WriteTo(&buf, 0); err != nil {
		return nil, err
	}

	p, err := profile.Parse(&buf)
	if err != nil {
		return nil, err
	}

	lines := parsePProfBlockTop(p, count)

	var out bytes.Buffer
	for _, line := range lines {
		out.WriteString(line + "\n")
	}

	return out.Bytes(), nil
}

func parsePProfBlockTop(p *profile.Profile, topN int) []string {
	flatDelay := map[string]time.Duration{}
	cumDelay := map[string]time.Duration{}
	flatCount := map[string]int64{}
	cumCount := map[string]int64{}
	var totalDelay time.Duration

	for _, s := range p.Sample {
		count := s.Value[0]
		delay := time.Duration(s.Value[1]) * time.Nanosecond
		totalDelay += delay

		seenInSample := map[string]bool{}
		for i, loc := range s.Location {
			for _, line := range loc.Line {
				fn := line.Function.Name

				if i == 0 {
					flatDelay[fn] += delay
					flatCount[fn] += count
				}

				if !seenInSample[fn] {
					cumDelay[fn] += delay
					cumCount[fn] += count
					seenInSample[fn] = true
				}
			}
		}
	}

	stats := make([]blockStat, 0, len(cumDelay))
	for fn, cd := range cumDelay {
		stats = append(stats, blockStat{
			Name:      fn,
			FlatDelay: flatDelay[fn],
			CumDelay:  cd,
			FlatCount: flatCount[fn],
			CumCount:  cumCount[fn],
		})
	}

	const nodeFraction = 0.005
	minCumTime := time.Duration(float64(totalDelay) * nodeFraction)

	filteredStats := make([]blockStat, 0, len(stats))
	for _, s := range stats {
		if s.CumDelay >= minCumTime {
			filteredStats = append(filteredStats, s)
		}
	}
	droppedNodesCount := len(stats) - len(filteredStats)

	sort.Slice(filteredStats, func(i, j int) bool {
		if filteredStats[i].FlatDelay != filteredStats[j].FlatDelay {
			return filteredStats[i].FlatDelay > filteredStats[j].FlatDelay
		}
		if filteredStats[i].CumDelay != filteredStats[j].CumDelay {
			return filteredStats[i].CumDelay > filteredStats[j].CumDelay
		}
		return filteredStats[i].Name < filteredStats[j].Name
	})

	if topN < 0 || topN > len(filteredStats) {
		topN = len(filteredStats)
	}

	out := []string{
		fmt.Sprintf("Showing nodes accounting for %s, 100%% of %s total",
			formatDuration(totalDelay), formatDuration(totalDelay)),
	}
	if droppedNodesCount > 0 {
		out = append(out, fmt.Sprintf("Dropped %d nodes (cum <= %s)",
			droppedNodesCount, formatDuration(minCumTime)))
	}
	out = append(out, fmt.Sprintf("Showing top %d nodes out of %d", topN, len(filteredStats)))
	out = append(out, "      flat  flat%   sum%        cum   cum%")

	var cumSumDelay time.Duration
	for i := 0; i < topN; i++ {
		s := filteredStats[i]
		cumSumDelay += s.FlatDelay

		flatValStr := formatDuration(s.FlatDelay)
		cumValStr := formatDuration(s.CumDelay)

		var flatPct, sumPct, cumPct float64
		if totalDelay > 0 {
			flatPct = float64(s.FlatDelay) * 100 / float64(totalDelay)
			sumPct = float64(cumSumDelay) * 100 / float64(totalDelay)
			cumPct = float64(s.CumDelay) * 100 / float64(totalDelay)
		}

		out = append(out, fmt.Sprintf(
			"  %9s %6.2f%% %6.2f%%   %9s %6.2f%%  %s",
			flatValStr,
			flatPct,
			sumPct,
			cumValStr,
			cumPct,
			s.Name,
		))
	}
	return out
}

func formatDuration(d time.Duration) string {
	if d == 0 {
		return "0"
	}
	if d < time.Microsecond {
		return fmt.Sprintf("%.2fns", float64(d.Nanoseconds()))
	}
	if d < time.Millisecond {
		return fmt.Sprintf("%.2fus", d.Seconds()*1e6)
	}
	if d < time.Second {
		return fmt.Sprintf("%.2fms", d.Seconds()*1e3)
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}
