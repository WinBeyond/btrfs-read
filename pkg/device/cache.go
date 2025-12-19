package device

import (
	"container/list"
	"sync"
)

// BlockCache is an LRU block cache.
type BlockCache struct {
	mu       sync.RWMutex
	capacity int
	cache    map[uint64]*list.Element
	lruList  *list.List
}

type cacheEntry struct {
	key  uint64
	data []byte
}

// NewBlockCache creates a cache.
func NewBlockCache(capacity int) *BlockCache {
	if capacity <= 0 {
		capacity = 256 // Default cache of 256 blocks.
	}

	return &BlockCache{
		capacity: capacity,
		cache:    make(map[uint64]*list.Element),
		lruList:  list.New(),
	}
}

// Get retrieves a cached entry.
func (c *BlockCache) Get(key uint64) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.cache[key]; ok {
		// Move to front (most recently used).
		c.lruList.MoveToFront(elem)
		entry := elem.Value.(*cacheEntry)
		// Return a copy to avoid external mutation.
		dataCopy := make([]byte, len(entry.data))
		copy(dataCopy, entry.data)
		return dataCopy, true
	}

	return nil, false
}

// Put adds a cached entry.
func (c *BlockCache) Put(key uint64, data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If it exists, update and move to the front.
	if elem, ok := c.cache[key]; ok {
		c.lruList.MoveToFront(elem)
		entry := elem.Value.(*cacheEntry)
		// Copy data.
		entry.data = make([]byte, len(data))
		copy(entry.data, data)
		return
	}

	// Create a new entry.
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)
	entry := &cacheEntry{key: key, data: dataCopy}
	elem := c.lruList.PushFront(entry)
	c.cache[key] = elem

	// Evict the oldest if over capacity.
	if c.lruList.Len() > c.capacity {
		c.evictOldest()
	}
}

// evictOldest removes the oldest cached entry.
func (c *BlockCache) evictOldest() {
	oldest := c.lruList.Back()
	if oldest != nil {
		c.lruList.Remove(oldest)
		entry := oldest.Value.(*cacheEntry)
		delete(c.cache, entry.key)
	}
}

// Clear clears the cache.
func (c *BlockCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[uint64]*list.Element)
	c.lruList = list.New()
}

// Len returns the number of cached entries.
func (c *BlockCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.lruList.Len()
}

// Capacity returns the cache capacity.
func (c *BlockCache) Capacity() int {
	return c.capacity
}

// CacheStats holds cache statistics.
type CacheStats struct {
	Size     int
	Capacity int
}

// Stats returns cache statistics.
func (c *BlockCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return CacheStats{
		Size:     c.lruList.Len(),
		Capacity: c.capacity,
	}
}
