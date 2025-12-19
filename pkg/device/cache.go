package device

import (
	"container/list"
	"sync"
)

// BlockCache LRU 块缓存
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

// NewBlockCache 创建缓存
func NewBlockCache(capacity int) *BlockCache {
	if capacity <= 0 {
		capacity = 256 // 默认缓存 256 个块
	}

	return &BlockCache{
		capacity: capacity,
		cache:    make(map[uint64]*list.Element),
		lruList:  list.New(),
	}
}

// Get 获取缓存项
func (c *BlockCache) Get(key uint64) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.cache[key]; ok {
		// 移到最前面（最近使用）
		c.lruList.MoveToFront(elem)
		entry := elem.Value.(*cacheEntry)
		// 返回副本，避免外部修改
		dataCopy := make([]byte, len(entry.data))
		copy(dataCopy, entry.data)
		return dataCopy, true
	}

	return nil, false
}

// Put 添加缓存项
func (c *BlockCache) Put(key uint64, data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 如果已存在，更新并移到前面
	if elem, ok := c.cache[key]; ok {
		c.lruList.MoveToFront(elem)
		entry := elem.Value.(*cacheEntry)
		// 复制数据
		entry.data = make([]byte, len(data))
		copy(entry.data, data)
		return
	}

	// 创建新条目
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)
	entry := &cacheEntry{key: key, data: dataCopy}
	elem := c.lruList.PushFront(entry)
	c.cache[key] = elem

	// 如果超过容量，删除最旧的
	if c.lruList.Len() > c.capacity {
		c.evictOldest()
	}
}

// evictOldest 删除最旧的缓存项
func (c *BlockCache) evictOldest() {
	oldest := c.lruList.Back()
	if oldest != nil {
		c.lruList.Remove(oldest)
		entry := oldest.Value.(*cacheEntry)
		delete(c.cache, entry.key)
	}
}

// Clear 清空缓存
func (c *BlockCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[uint64]*list.Element)
	c.lruList = list.New()
}

// Len 返回当前缓存项数量
func (c *BlockCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.lruList.Len()
}

// Capacity 返回缓存容量
func (c *BlockCache) Capacity() int {
	return c.capacity
}

// Stats 返回缓存统计信息
type CacheStats struct {
	Size     int
	Capacity int
}

// Stats 获取缓存统计
func (c *BlockCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return CacheStats{
		Size:     c.lruList.Len(),
		Capacity: c.capacity,
	}
}
