package chunk

import (
	"encoding/binary"
	"fmt"

	"github.com/WinBeyond/btrfs-read/pkg/btree"
	"github.com/WinBeyond/btrfs-read/pkg/logger"
)

// ChunkTreeLoader loads all chunks from the chunk tree.
type ChunkTreeLoader struct {
	manager  *Manager
	reader   btree.NodeReader
	nodeSize uint32
}

// NewChunkTreeLoader creates a loader.
func NewChunkTreeLoader(manager *Manager, reader btree.NodeReader, nodeSize uint32) *ChunkTreeLoader {
	return &ChunkTreeLoader{
		manager:  manager,
		reader:   reader,
		nodeSize: nodeSize,
	}
}

// LoadFromChunkTree loads all chunks from the chunk tree.
func (l *ChunkTreeLoader) LoadFromChunkTree(chunkRoot uint64) error {
	// Recursively read the entire chunk tree.
	return l.readChunkNode(chunkRoot)
}

// readChunkNode recursively reads chunk tree nodes.
func (l *ChunkTreeLoader) readChunkNode(nodeAddr uint64) error {
	// Read node.
	node, err := l.reader.ReadNode(nodeAddr, l.nodeSize)
	if err != nil {
		return fmt.Errorf("failed to read node at 0x%x: %w", nodeAddr, err)
	}

	// If it's a leaf node, parse all chunk items.
	if node.Header.Level == 0 {
		for _, item := range node.Items {
			// Only handle CHUNK_ITEM (objectid=256, type=228).
			if item.Key.ObjectID == 256 && item.Key.Type == 228 {
				if err := l.parseAndAddChunk(item.Key.Offset, item.Data); err != nil {
					// Log the error only; do not stop loading (may be unsupported RAID types).
					logger.Warn("Failed to parse chunk at offset 0x%x: %v", item.Key.Offset, err)
				}
			}
		}
		return nil
	}

	// If it's an internal node, recursively read all child nodes.
	for _, ptr := range node.Ptrs {
		if err := l.readChunkNode(ptr); err != nil {
			return err
		}
	}

	return nil
}

// parseAndAddChunk parses and adds a chunk.
func (l *ChunkTreeLoader) parseAndAddChunk(logicalOffset uint64, data []byte) error {
	if len(data) < 50 {
		return fmt.Errorf("chunk data too short: %d bytes", len(data))
	}

	// Parse chunk item (header 48 bytes).
	length := binary.LittleEndian.Uint64(data[0:8])
	chunkType := binary.LittleEndian.Uint64(data[24:32])
	numStripes := binary.LittleEndian.Uint16(data[44:46]) // offset 44, not 48

	// Check RAID type (simplified: only SINGLE/DUP).
	raidBits := chunkType & 0x1F8 // bits 3-8

	if raidBits != 0 && raidBits != (1<<5) { // Allow DUP (1<<5).
		// Skip complex RAID types.
		return nil
	}

	if numStripes < 1 {
		return fmt.Errorf("invalid num_stripes: %d", numStripes)
	}

	// Read the first stripe (starts at offset 48).
	stripeOffset := 48
	if stripeOffset+32 > len(data) {
		return fmt.Errorf("stripe data out of bounds")
	}

	devID := binary.LittleEndian.Uint64(data[stripeOffset:])
	physicalOffset := binary.LittleEndian.Uint64(data[stripeOffset+8:])

	// Create chunk mapping.
	mapping := &ChunkMapping{
		LogicalStart:  logicalOffset,
		LogicalLength: length,
		PhysicalStart: physicalOffset,
		DeviceID:      devID,
	}

	l.manager.AddMapping(mapping)

	return nil
}
