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

// Manager manages chunks (simplified: uses slices instead of a red-black tree).
type Manager struct {
	mu       sync.RWMutex
	mappings []*ChunkMapping // Sorted by LogicalStart.
}

// NewManager creates a manager.
func NewManager() *Manager {
	return &Manager{
		mappings: make([]*ChunkMapping, 0),
	}
}

// AddMapping adds a chunk mapping.
func (m *Manager) AddMapping(mapping *ChunkMapping) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.mappings = append(m.mappings, mapping)

	// Keep sorted.
	sort.Slice(m.mappings, func(i, j int) bool {
		return m.mappings[i].LogicalStart < m.mappings[j].LogicalStart
	})
}

// LogicalToPhysical maps a logical address to a physical address.
func (m *Manager) LogicalToPhysical(logical uint64) (*PhysicalAddr, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Binary search for the chunk containing the logical address.
	mapping := m.findMapping(logical)
	if mapping == nil {
		return nil, errors.Wrap("LogicalToPhysical",
			fmt.Errorf("no chunk mapping found for logical 0x%x", logical))
	}

	// Map address.
	return mapping.MapAddress(logical)
}

func (m *Manager) findMapping(logical uint64) *ChunkMapping {
	// Binary search.
	idx := sort.Search(len(m.mappings), func(i int) bool {
		return m.mappings[i].LogicalStart > logical
	})

	// idx is the first start > logical index.
	// We need to check the previous entry.
	if idx > 0 {
		candidate := m.mappings[idx-1]
		if candidate.Contains(logical) {
			return candidate
		}
	}

	return nil
}

// ParseSystemChunkArray parses chunks from the superblock system chunk array.
func (m *Manager) ParseSystemChunkArray(data []byte, arraySize uint32) error {

	if arraySize == 0 {
		logger.Warn("System chunk array size is 0")
		return nil
	}

	offset := 0
	for offset < int(arraySize) {
		// Ensure enough space to read a minimal chunk (key + header + 1 stripe).
		minChunkSize := 17 + 50 + 32 // 99 bytes
		if offset+minChunkSize > int(arraySize) {
			// Not enough remaining space; stop parsing.
			break
		}

		// Parse key (17 bytes).
		objectID := binary.LittleEndian.Uint64(data[offset:])
		keyType := data[offset+8]
		keyOffset := binary.LittleEndian.Uint64(data[offset+9:])
		offset += 17

		// Check for a valid CHUNK_ITEM.
		// objectID should be FIRST_CHUNK_TREE_OBJECTID (256).
		// keyType should be CHUNK_ITEM_KEY (228).
		if objectID != 256 || keyType != 228 {
			// Not a chunk item; this should not happen.
			return fmt.Errorf("invalid system chunk array entry: objectid=%d, type=%d", objectID, keyType)
		}

		// Ensure enough space to read the chunk header.
		if offset+50 > int(arraySize) {
			return fmt.Errorf("chunk header exceeds array size")
		}

		// Parse chunk item header (48 bytes).
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

		// Ensure enough space to read stripes.
		if offset+int(numStripes)*32 > int(arraySize) {
			logger.Error("Stripe data exceeds array size: need %d bytes, only %d remaining (arraySize=%d, offset=%d)",
				int(numStripes)*32, int(arraySize)-offset, arraySize, offset)
			return fmt.Errorf("stripe data exceeds array size")
		}

		// Only SINGLE/DUP types are supported (simplified).
		raidBits := chunkType & 0x1F8
		if raidBits != 0 && raidBits != (1<<5) {
			offset += int(numStripes) * 32
			continue
		}

		if numStripes < 1 {
			return fmt.Errorf("invalid num_stripes: %d", numStripes)
		}

		// Read the first stripe.
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

// Len returns the number of chunks.
func (m *Manager) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.mappings)
}
