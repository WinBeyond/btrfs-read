package chunk

import (
	"encoding/binary"
	"fmt"
	"sort"
	"sync"

	"github.com/WinBeyond/btrfs-read/pkg/errors"
	"github.com/WinBeyond/btrfs-read/pkg/logger"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Manager Chunk 管理器（简化版：使用切片而非红黑树）
type Manager struct {
	mu       sync.RWMutex
	mappings []*ChunkMapping // 按 LogicalStart 排序
}

// NewManager 创建管理器
func NewManager() *Manager {
	return &Manager{
		mappings: make([]*ChunkMapping, 0),
	}
}

// AddMapping 添加 Chunk 映射
func (m *Manager) AddMapping(mapping *ChunkMapping) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.mappings = append(m.mappings, mapping)

	// 保持排序
	sort.Slice(m.mappings, func(i, j int) bool {
		return m.mappings[i].LogicalStart < m.mappings[j].LogicalStart
	})
}

// LogicalToPhysical 逻辑地址转物理地址
func (m *Manager) LogicalToPhysical(logical uint64) (*PhysicalAddr, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 二分查找包含该逻辑地址的 chunk
	mapping := m.findMapping(logical)
	if mapping == nil {
		return nil, errors.Wrap("LogicalToPhysical",
			fmt.Errorf("no chunk mapping found for logical 0x%x", logical))
	}

	// 映射地址
	return mapping.MapAddress(logical)
}

func (m *Manager) findMapping(logical uint64) *ChunkMapping {
	// 二分查找
	idx := sort.Search(len(m.mappings), func(i int) bool {
		return m.mappings[i].LogicalStart > logical
	})

	// idx 是第一个 start > logical 的索引
	// 我们需要检查前一个
	if idx > 0 {
		candidate := m.mappings[idx-1]
		if candidate.Contains(logical) {
			return candidate
		}
	}

	return nil
}

// ParseSystemChunkArray 从 superblock 的系统 chunk 数组解析 chunk
func (m *Manager) ParseSystemChunkArray(data []byte, arraySize uint32) error {

	if arraySize == 0 {
		logger.Warn("System chunk array size is 0")
		return nil
	}

	offset := 0
	for offset < int(arraySize) {
		// 确保有足够的空间读取最小chunk（key + header + 1 stripe）
		minChunkSize := 17 + 50 + 32 // 99 bytes
		if offset+minChunkSize > int(arraySize) {
			// 剩余空间不足，结束解析
			break
		}

		// 解析 key (17 bytes)
		objectID := binary.LittleEndian.Uint64(data[offset:])
		keyType := data[offset+8]
		keyOffset := binary.LittleEndian.Uint64(data[offset+9:])
		offset += 17

		// 检查是否是有效的 CHUNK_ITEM
		// objectID 应该是 FIRST_CHUNK_TREE_OBJECTID (256)
		// keyType 应该是 CHUNK_ITEM_KEY (228)
		if objectID != 256 || keyType != 228 {
			// 不是 chunk item，这不应该发生
			return fmt.Errorf("invalid system chunk array entry: objectid=%d, type=%d", objectID, keyType)
		}

		// 确保有足够空间读取 chunk header
		if offset+50 > int(arraySize) {
			return fmt.Errorf("chunk header exceeds array size")
		}

		// 解析 chunk item header (48 bytes)
		// struct btrfs_chunk {
		//   __le64 length;
		//   __le64 __unused1;
		//   __le64 stripe_len;
		//   __le64 type;
		//   __le32 __unused2[3];  // 12 bytes
		//   __le16 num_stripes;
		//   __le16 sub_stripes;
		// }
		length := binary.LittleEndian.Uint64(data[offset:])
		chunkType := binary.LittleEndian.Uint64(data[offset+24:])
		numStripes := binary.LittleEndian.Uint16(data[offset+44:])
		offset += 48

		// 确保有足够空间读取 stripes
		if offset+int(numStripes)*32 > int(arraySize) {
			logger.Error("Stripe data exceeds array size: need %d bytes, only %d remaining (arraySize=%d, offset=%d)",
				int(numStripes)*32, int(arraySize)-offset, arraySize, offset)
			return fmt.Errorf("stripe data exceeds array size")
		}

		// 只支持 SINGLE/DUP 类型（简化）
		raidBits := chunkType & 0x1F8
		if raidBits != 0 && raidBits != (1<<5) {
			offset += int(numStripes) * 32
			continue
		}

		if numStripes < 1 {
			return fmt.Errorf("invalid num_stripes: %d", numStripes)
		}

		// 读取第一个 stripe
		devID := binary.LittleEndian.Uint64(data[offset:])
		stripeOffset := binary.LittleEndian.Uint64(data[offset+8:])

		mapping := &ChunkMapping{
			LogicalStart:  keyOffset,
			LogicalLength: length,
			PhysicalStart: stripeOffset,
			DeviceID:      devID,
		}

		m.AddMapping(mapping)

		offset += int(numStripes) * 32
	}

	return nil
}

// Len 返回 chunk 数量
func (m *Manager) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.mappings)
}
