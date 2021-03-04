package fzf

import (
	"fmt"
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/junegunn/fzf/src/util"
)

// MatchRequest represents a search request
type MatchRequest struct {
	chunks     []*Chunk
	pattern    *Pattern
	final      bool
	sort       bool
	clearCache bool
}

// Matcher is responsible for performing search
type Matcher struct {
	patternBuilder func([]rune) *Pattern
	sort           bool
	tac            bool
	eventBox       *util.EventBox
	reqBox         *util.EventBox
	partitions     int
	slab           []*util.Slab
	mergerCache    map[string]*Merger
}

const (
	reqRetry util.EventType = iota
	reqReset
)

// NewMatcher returns a new Matcher
func NewMatcher(patternBuilder func([]rune) *Pattern,
	sort bool, tac bool, eventBox *util.EventBox) *Matcher {
	partitions := util.Min(numPartitionsMultiplier*runtime.NumCPU(), maxPartitions)
	return &Matcher{
		patternBuilder: patternBuilder,
		sort:           sort,
		tac:            tac,
		eventBox:       eventBox,
		reqBox:         util.NewEventBox(),
		partitions:     partitions,
		slab:           make([]*util.Slab, partitions),
		mergerCache:    make(map[string]*Merger)}
}

// Loop puts Matcher in action
func (m *Matcher) Loop() {
	prevCount := 0

	for {
		var request MatchRequest

		m.reqBox.Wait(func(events *util.Events) {
			for _, val := range *events {
				switch val := val.(type) {
				case MatchRequest:
					request = val
				default:
					panic(fmt.Sprintf("Unexpected type: %T", val))
				}
			}
			events.Clear()
		})

		if request.sort != m.sort || request.clearCache {
			m.sort = request.sort
			m.mergerCache = make(map[string]*Merger)
			clearChunkCache()
		}

		// Restart search
		patternString := request.pattern.AsString()
		var merger *Merger
		cancelled := false
		count := CountItems(request.chunks)

		foundCache := false
		if count == prevCount {
			// Look up mergerCache
			if cached, found := m.mergerCache[patternString]; found {
				foundCache = true
				merger = cached
			}
		} else {
			// Invalidate mergerCache
			prevCount = count
			m.mergerCache = make(map[string]*Merger)
		}

		if !foundCache {
			merger, cancelled = m.scan(request)
		}

		if !cancelled {
			if merger.cacheable() {
				m.mergerCache[patternString] = merger
			}
			merger.final = request.final
			m.eventBox.Set(EvtSearchFin, merger)
		}
	}
}

func (m *Matcher) sliceChunks(chunks []*Chunk) [][]*Chunk {
	partitions := m.partitions
	perSlice := len(chunks) / partitions

	if perSlice == 0 {
		partitions = len(chunks)
		perSlice = 1
	}

	slices := make([][]*Chunk, partitions)
	for i := 0; i < partitions; i++ {
		start := i * perSlice
		end := start + perSlice
		if i == partitions-1 {
			end = len(chunks)
		}
		slices[i] = chunks[start:end]
	}
	return slices
}

func (m *Matcher) scan(request MatchRequest) (*Merger, bool) {
	startedAt := time.Now()

	numChunks := len(request.chunks)
	if numChunks == 0 {
		return EmptyMerger, false
	}
	pattern := request.pattern
	if pattern.IsEmpty() {
		return PassMerger(&request.chunks, m.tac), false
	}

	cancelled := util.NewAtomicBool(false)
	defer cancelled.Set(true)

	slices := m.sliceChunks(request.chunks)
	matchSema := make(chan struct{}, numChunks)
	waitGroup := sync.WaitGroup{}

	partialResults := make([][]Result, len(slices))

	for idx, chunks := range slices {
		waitGroup.Add(1)
		if m.slab[idx] == nil {
			m.slab[idx] = util.MakeSlab(slab16Size, slab32Size)
		}
		go func(idx int, slab *util.Slab, chunks []*Chunk) {
			defer waitGroup.Done()
			var all []Result
			for _, chunk := range chunks {
				matches := request.pattern.Match(chunk, slab)
				if cancelled.Get() {
					return
				}
				all = append(all, matches...)
				matchSema <- struct{}{}
			}
			if m.sort {
				if m.tac {
					sort.Sort(ByRelevanceTac(all))
				} else {
					sort.Sort(ByRelevance(all))
				}
			}
			partialResults[idx] = all
		}(idx, m.slab[idx], chunks)
	}

	count := 0
	for range matchSema {
		count++
		if count == numChunks {
			break
		}

		if m.reqBox.Peek(reqReset) {
			cancelled.Set(true)
			waitGroup.Wait()
			return nil, true
		}

		if time.Since(startedAt) > progressMinDuration {
			m.eventBox.Set(EvtSearchProgress, float32(count)/float32(numChunks))
		}
	}

	waitGroup.Wait()
	return NewMerger(pattern, partialResults, m.sort, m.tac), false
}

// Reset is called to interrupt/signal the ongoing search
func (m *Matcher) Reset(chunks []*Chunk, patternRunes []rune, cancel bool, final bool, sort bool, clearCache bool) {
	pattern := m.patternBuilder(patternRunes)

	var event util.EventType
	if cancel {
		event = reqReset
	} else {
		event = reqRetry
	}
	m.reqBox.Set(event, MatchRequest{chunks, pattern, final, sort && pattern.sortable, clearCache})
}
