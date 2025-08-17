package OOD

import "sync"

type SafeCache struct {
	cache map[string]interface{}
	mu    sync.RWMutex
}

func NewSafeCache() *SafeCache {
	return &SafeCache{
		cache: make(map[string]interface{}),
	}
}

func (sc *SafeCache) Set(key string, value interface{}) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.cache[key] = value
}

func (sc *SafeCache) Get(key string) (interface{}, bool) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	value, ok := sc.cache[key]
	return value, ok
}

func (sc *SafeCache) Delete(key string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	delete(sc.cache, key)
}
