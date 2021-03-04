package fzf

import "sync"

// queryCache associates strings to lists of items
type queryCache map[string][]Result

// ChunkCache associates Chunk and query string to lists of items
type ChunkCache struct {
	mutex sync.RWMutex
	cache map[*Chunk]*queryCache
}

// NewChunkCache returns a new ChunkCache
func NewChunkCache() ChunkCache {
	return ChunkCache{cache: make(map[*Chunk]*queryCache)}
}

// Add adds the list to the cache
func (cc *ChunkCache) Add(chunk *Chunk, key string, list []Result) {
	if len(key) == 0 || !chunk.IsFull() || len(list) > queryCacheMax {
		return
	}

	cc.mutex.Lock()
	qc, ok := cc.cache[chunk]
	if !ok {
		qc = &queryCache{}
		cc.cache[chunk] = qc
	}
	(*qc)[key] = list
	cc.mutex.Unlock()
}

// Lookup is called to lookup ChunkCache
func (cc *ChunkCache) Lookup(chunk *Chunk, key string) (result []Result) {
	if len(key) == 0 || !chunk.IsFull() {
		return nil
	}

	cc.mutex.RLock()
	if qc, ok := cc.cache[chunk]; ok {
		result = (*qc)[key]
	}
	cc.mutex.RUnlock()
	return result
}

func (cc *ChunkCache) Search(chunk *Chunk, key string) []Result {
	if len(key) == 0 || !chunk.IsFull() {
		return nil
	}

	cc.mutex.RLock()
	if qc, ok := cc.cache[chunk]; ok {
		for idx := 1; idx < len(key); idx++ {
			// [---------| ] | [ |---------]
			// [--------|  ] | [  |--------]
			// [-------|   ] | [   |-------]
			prefix := key[:len(key)-idx]
			suffix := key[idx:]
			for _, substr := range [2]string{prefix, suffix} {
				if cached, found := (*qc)[substr]; found {
					cc.mutex.RUnlock()
					return cached
				}
			}
		}
	}
	cc.mutex.RUnlock()

	return nil
}
