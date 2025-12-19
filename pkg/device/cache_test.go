package device

import (
	"testing"
)

func TestBlockCache_Basic(t *testing.T) {
	cache := NewBlockCache(3)

	// 测试基本的 Put 和 Get
	cache.Put(1, []byte("data1"))
	cache.Put(2, []byte("data2"))
	cache.Put(3, []byte("data3"))

	// 验证容量
	if cache.Len() != 3 {
		t.Errorf("Expected cache size 3, got %d", cache.Len())
	}

	// 测试 Get
	data, ok := cache.Get(1)
	if !ok {
		t.Error("Expected to find key 1")
	}
	if string(data) != "data1" {
		t.Errorf("Expected 'data1', got '%s'", string(data))
	}

	// 测试不存在的 key
	_, ok = cache.Get(999)
	if ok {
		t.Error("Expected key 999 not found")
	}
}

func TestBlockCache_LRU(t *testing.T) {
	cache := NewBlockCache(2)

	// 添加两个条目
	cache.Put(1, []byte("data1"))
	cache.Put(2, []byte("data2"))

	// 访问第一个（使其变为最近使用）
	cache.Get(1)

	// 添加第三个，应该淘汰 key=2
	cache.Put(3, []byte("data3"))

	// 验证 key=2 被淘汰
	_, ok := cache.Get(2)
	if ok {
		t.Error("Expected key 2 to be evicted")
	}

	// 验证 key=1 和 key=3 仍然存在
	_, ok = cache.Get(1)
	if !ok {
		t.Error("Expected key 1 to exist")
	}

	_, ok = cache.Get(3)
	if !ok {
		t.Error("Expected key 3 to exist")
	}
}

func TestBlockCache_Update(t *testing.T) {
	cache := NewBlockCache(2)

	// 添加条目
	cache.Put(1, []byte("data1"))

	// 更新同一个 key
	cache.Put(1, []byte("updated"))

	// 验证数据被更新
	data, ok := cache.Get(1)
	if !ok {
		t.Error("Expected to find key 1")
	}
	if string(data) != "updated" {
		t.Errorf("Expected 'updated', got '%s'", string(data))
	}

	// 验证缓存大小没有增加
	if cache.Len() != 1 {
		t.Errorf("Expected cache size 1, got %d", cache.Len())
	}
}

func TestBlockCache_Clear(t *testing.T) {
	cache := NewBlockCache(10)

	// 添加多个条目
	for i := 0; i < 5; i++ {
		cache.Put(uint64(i), []byte{byte(i)})
	}

	if cache.Len() != 5 {
		t.Errorf("Expected cache size 5, got %d", cache.Len())
	}

	// 清空缓存
	cache.Clear()

	if cache.Len() != 0 {
		t.Errorf("Expected cache size 0, got %d", cache.Len())
	}

	// 验证所有数据都被清除
	for i := 0; i < 5; i++ {
		_, ok := cache.Get(uint64(i))
		if ok {
			t.Errorf("Expected key %d to be cleared", i)
		}
	}
}

func TestBlockCache_DataIsolation(t *testing.T) {
	cache := NewBlockCache(10)

	// 原始数据
	original := []byte("original")
	cache.Put(1, original)

	// 修改原始数据
	original[0] = 'X'

	// 获取缓存数据
	cached, ok := cache.Get(1)
	if !ok {
		t.Error("Expected to find key 1")
	}

	// 验证缓存数据未被修改
	if string(cached) != "original" {
		t.Errorf("Expected 'original', got '%s'", string(cached))
	}

	// 修改获取的数据
	cached[0] = 'Y'

	// 再次获取，验证缓存未被修改
	cached2, _ := cache.Get(1)
	if string(cached2) != "original" {
		t.Errorf("Expected 'original', got '%s'", string(cached2))
	}
}

func TestBlockCache_Stats(t *testing.T) {
	cache := NewBlockCache(5)

	// 初始状态
	stats := cache.Stats()
	if stats.Size != 0 {
		t.Errorf("Expected initial size 0, got %d", stats.Size)
	}
	if stats.Capacity != 5 {
		t.Errorf("Expected capacity 5, got %d", stats.Capacity)
	}

	// 添加数据
	cache.Put(1, []byte("data"))
	stats = cache.Stats()
	if stats.Size != 1 {
		t.Errorf("Expected size 1, got %d", stats.Size)
	}
}

func BenchmarkBlockCache_Put(b *testing.B) {
	cache := NewBlockCache(1000)
	data := make([]byte, 4096) // 4KB

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Put(uint64(i%1000), data)
	}
}

func BenchmarkBlockCache_Get(b *testing.B) {
	cache := NewBlockCache(1000)
	data := make([]byte, 4096)

	// 预填充缓存
	for i := 0; i < 1000; i++ {
		cache.Put(uint64(i), data)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(uint64(i % 1000))
	}
}

func BenchmarkBlockCache_Mixed(b *testing.B) {
	cache := NewBlockCache(1000)
	data := make([]byte, 4096)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i%2 == 0 {
			cache.Put(uint64(i%1000), data)
		} else {
			cache.Get(uint64(i % 1000))
		}
	}
}
