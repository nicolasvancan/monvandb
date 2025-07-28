package files

import (
	"sync"
	"time"

	"github.com/nicolasvancan/monvandb/src/btree"
)

type MappedPages = map[uint64]btree.TreeNode

// PageCache is a structure that holds a cache of pages.
// It uses a map to store the cached pages, where the key is a string identifier
// and the value is a byte slice representing the page content.
type PageCache struct {
	MappedPages MappedPages
	Mutex       sync.RWMutex
}

func NewPageCache() *PageCache {
	return &PageCache{
		MappedPages: make(MappedPages),
		Mutex:       sync.RWMutex{},
	}
}

// Get retrieves a page from the cache by its identifier.
// It returns the page content and a boolean indicating whether the page was found.
func (pc *PageCache) Get(id uint64) (btree.TreeNode, bool) {
	pc.Mutex.RLock()

	page, exists := pc.MappedPages[id]

	defer pc.Mutex.RUnlock()
	return page, exists
}

// Set adds or updates a page in the cache.
func (pc *PageCache) Set(id uint64, page btree.TreeNode) {
	pc.Mutex.Lock()

	pc.MappedPages[id] = page
	defer pc.Mutex.Unlock()
}

// Delete removes a page from the cache by its identifier.
func (pc *PageCache) Delete(id uint64) {
	pc.Mutex.Lock()

	delete(pc.MappedPages, id)
	defer pc.Mutex.Unlock()
}

// Clear removes all pages from the cache.
func (pc *PageCache) Clear() {
	pc.Mutex.Lock()

	pc.MappedPages = make(MappedPages)
	defer pc.Mutex.Unlock()
}

// Len returns the number of pages currently in the cache.
func (pc *PageCache) Len() int {
	pc.Mutex.RLock()

	length := len(pc.MappedPages)
	defer pc.Mutex.RUnlock()
	return length
}

// Keys returns a slice of all keys in the cache.
func (pc *PageCache) Keys() []uint64 {
	pc.Mutex.RLock()

	keys := make([]uint64, 0, len(pc.MappedPages))
	for key := range pc.MappedPages {
		keys = append(keys, key)
	}

	defer pc.Mutex.RUnlock()
	return keys
}

/*
CacheSystem is a structure that represents a cache system for a index or main database file, nameless BTreeFile.
It is used to manage cached pages that are already loaded into memory.

It contains a PageCache to store cached pages of type PageCache, which allows for efficient retrieval and storage of page data.
*/
type CacheSystem struct {
	Ttl          int
	PageCache    *PageCache
	Header       []byte
	ToBeUpdated  []uint64
	ToBeInserted []uint64
	LastAccessed map[uint64]time.Time
	MaxCacheSize int // Maximum size of the cache, can be used to limit the number of cached pages.
	Mutex        sync.RWMutex
}

func NewCacheSystem(ttl int, maxCacheSize int) *CacheSystem {
	return &CacheSystem{
		Ttl:          ttl,
		PageCache:    NewPageCache(),
		ToBeUpdated:  make([]uint64, 0),
		LastAccessed: make(map[uint64]time.Time),
		MaxCacheSize: maxCacheSize,
		Mutex:        sync.RWMutex{},
	}
}

func ClearTimedOutCache(cache *CacheSystem) {

	ticker := time.NewTicker(time.Duration(cache.Ttl) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		cache.Mutex.Lock()
		for id, lastAccessed := range cache.LastAccessed {
			if now.Sub(lastAccessed) > time.Duration(cache.Ttl)*time.Second {
				cache.PageCache.Delete(id)
				delete(cache.LastAccessed, id)
			}
		}
		cache.Mutex.Unlock()
	}
}

func (cs *CacheSystem) Get(id uint64) (btree.TreeNode, bool) {
	// PageCache.Get has its own RLock.
	page, exists := cs.PageCache.Get(id)

	if exists {
		cs.Mutex.Lock()
		cs.LastAccessed[id] = time.Now()
		cs.Mutex.Unlock()
	}
	return page, exists
}
func findOldest(cache *CacheSystem) (uint64, time.Time) {
	oldestID := uint64(0)
	oldestTime := time.Now()

	cache.Mutex.Lock()
	for id, lastAccessed := range cache.LastAccessed {
		if lastAccessed.Before(oldestTime) {
			oldestTime = lastAccessed
			oldestID = id
		}
	}
	cache.Mutex.Unlock()

	return oldestID, oldestTime
}
func isCacheFull(cache *CacheSystem) bool {
	return cache.PageCache.Len()*btree.PAGE_SIZE >= cache.MaxCacheSize
}

func (cs *CacheSystem) GetHeader() []byte {
	cs.Mutex.RLock()
	defer cs.Mutex.RUnlock() // Ensure mutex is unlocked after reading header
	return cs.Header
}

func (cs *CacheSystem) SetHeader(header []byte) {
	cs.Mutex.Lock()
	cs.Header = header
	defer cs.Mutex.Unlock()
}

func (cs *CacheSystem) Set(id uint64, page btree.TreeNode) {
	cs.Mutex.Lock()
	if isCacheFull(cs) {
		// If the cache size exceeds the maximum, remove the oldest entry
		oldestID, _ := findOldest(cs)

		if oldestID != 0 {
			cs.PageCache.Delete(oldestID)
			delete(cs.LastAccessed, oldestID)
		}
	} else {
		cs.PageCache.Set(id, page)

	}
	cs.LastAccessed[id] = time.Now()
	cs.ToBeUpdated = append(cs.ToBeUpdated, id)
	defer cs.Mutex.Unlock()

}

func (cs *CacheSystem) Delete(id uint64) {
	cs.Mutex.Lock()
	cs.PageCache.Delete(id)
	delete(cs.LastAccessed, id)
	defer cs.Mutex.Unlock()
}

func (cs *CacheSystem) Create(id uint64, page btree.TreeNode) {
	cs.Mutex.Lock()

	if isCacheFull(cs) {
		// If the cache size exceeds the maximum, remove the oldest entry
		oldestID, _ := findOldest(cs)

		if oldestID != 0 {
			cs.PageCache.Delete(oldestID)
			delete(cs.LastAccessed, oldestID)
		}
	} else {
		cs.PageCache.Set(id, page)

	}
	cs.LastAccessed[id] = time.Now()
	cs.ToBeInserted = append(cs.ToBeInserted, id)
	defer cs.Mutex.Unlock()
}
