package device

import (
	"testing"
)

func TestBlockCache_Basic(t *testing.T) {
	cache := NewBlockCache(3)

	// Test basic Put and Get.
	cache.Put(1, []byte("data1"))
	cache.Put(2, []byte("data2"))
	cache.Put(3, []byte("data3"))

	// Verify capacity.
	if cache.Len() != 3 {
		t.Errorf("Expected cache size 3, got %d", cache.Len())
	}

	// Test Get.
	data, ok := cache.Get(1)
	if !ok {
		t.Error("Expected to find key 1")
	}
	if string(data) != "data1" {
		t.Errorf("Expected 'data1', got '%s'", string(data))
	}

	// Test non-existent key.
	_, ok = cache.Get(999)
	if ok {
		t.Error("Expected key 999 not found")
	}
}

func TestBlockCache_LRU(t *testing.T) {
	cache := NewBlockCache(2)

	// Add two entries.
	cache.Put(1, []byte("data1"))
	cache.Put(2, []byte("data2"))

	// Access the first (make it most recently used).
	cache.Get(1)

	// Add third entry; should evict key=2.
	cache.Put(3, []byte("data3"))

	// Verify key=2 was evicted.
	_, ok := cache.Get(2)
	if ok {
		t.Error("Expected key 2 to be evicted")
	}

	// Verify key=1 and key=3 still exist.
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

	// Add entry.
	cache.Put(1, []byte("data1"))

	// Update the same key.
	cache.Put(1, []byte("updated"))

	// Verify data updated.
	data, ok := cache.Get(1)
	if !ok {
		t.Error("Expected to find key 1")
	}
	if string(data) != "updated" {
		t.Errorf("Expected 'updated', got '%s'", string(data))
	}

	// Verify cache size did not grow.
	if cache.Len() != 1 {
		t.Errorf("Expected cache size 1, got %d", cache.Len())
	}
}

func TestBlockCache_Clear(t *testing.T) {
	cache := NewBlockCache(10)

	// Add multiple entries.
	for i := 0; i < 5; i++ {
		cache.Put(uint64(i), []byte{byte(i)})
	}

	if cache.Len() != 5 {
		t.Errorf("Expected cache size 5, got %d", cache.Len())
	}

	// Clear cache.
	cache.Clear()

	if cache.Len() != 0 {
		t.Errorf("Expected cache size 0, got %d", cache.Len())
	}

	// Verify all data cleared.
	for i := 0; i < 5; i++ {
		_, ok := cache.Get(uint64(i))
		if ok {
			t.Errorf("Expected key %d to be cleared", i)
		}
	}
}

func TestBlockCache_DataIsolation(t *testing.T) {
	cache := NewBlockCache(10)

	// Original data.
	original := []byte("original")
	cache.Put(1, original)

	// Modify original data.
	original[0] = 'X'

	// Get cached data.
	cached, ok := cache.Get(1)
	if !ok {
		t.Error("Expected to find key 1")
	}

	// Verify cached data unchanged.
	if string(cached) != "original" {
		t.Errorf("Expected 'original', got '%s'", string(cached))
	}

	// Modify retrieved data.
	cached[0] = 'Y'

	// Get again and verify cache unchanged.
	cached2, _ := cache.Get(1)
	if string(cached2) != "original" {
		t.Errorf("Expected 'original', got '%s'", string(cached2))
	}
}

func TestBlockCache_Stats(t *testing.T) {
	cache := NewBlockCache(5)

	// Initial state.
	stats := cache.Stats()
	if stats.Size != 0 {
		t.Errorf("Expected initial size 0, got %d", stats.Size)
	}
	if stats.Capacity != 5 {
		t.Errorf("Expected capacity 5, got %d", stats.Capacity)
	}

	// Add data.
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

	// Pre-fill cache.
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
