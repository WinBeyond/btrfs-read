package chunk

import (
	"encoding/binary"
	"fmt"

	"github.com/WinBeyond/btrfs-read/pkg/btree"
	"github.com/WinBeyond/btrfs-read/pkg/logger"
)

// ChunkTreeLoader 负责从 chunk tree 加载所有 chunk
type ChunkTreeLoader struct {
	manager  *Manager
	reader   btree.NodeReader
	nodeSize uint32
}

// NewChunkTreeLoader 创建加载器
func NewChunkTreeLoader(manager *Manager, reader btree.NodeReader, nodeSize uint32) *ChunkTreeLoader {
	return &ChunkTreeLoader{
		manager:  manager,
		reader:   reader,
		nodeSize: nodeSize,
	}
}

// LoadFromChunkTree 从 chunk tree 加载所有 chunk
func (l *ChunkTreeLoader) LoadFromChunkTree(chunkRoot uint64) error {
	// 递归读取整个 chunk tree
	return l.readChunkNode(chunkRoot)
}

// readChunkNode 递归读取 chunk tree 节点
func (l *ChunkTreeLoader) readChunkNode(nodeAddr uint64) error {
	// 读取节点
	node, err := l.reader.ReadNode(nodeAddr, l.nodeSize)
	if err != nil {
		return fmt.Errorf("failed to read node at 0x%x: %w", nodeAddr, err)
	}

	// 如果是叶子节点，解析所有 chunk items
	if node.Header.Level == 0 {
		for _, item := range node.Items {
			// 只处理 CHUNK_ITEM (objectid=256, type=228)
			if item.Key.ObjectID == 256 && item.Key.Type == 228 {
				if err := l.parseAndAddChunk(item.Key.Offset, item.Data); err != nil {
					// 只记录错误，不中断加载（可能是不支持的 RAID 类型）
					logger.Warn("Failed to parse chunk at offset 0x%x: %v", item.Key.Offset, err)
				}
			}
		}
		return nil
	}

	// 如果是内部节点，递归读取所有子节点
	for _, ptr := range node.Ptrs {
		if err := l.readChunkNode(ptr); err != nil {
			return err
		}
	}

	return nil
}

// parseAndAddChunk 解析并添加 chunk
func (l *ChunkTreeLoader) parseAndAddChunk(logicalOffset uint64, data []byte) error {
	if len(data) < 50 {
		return fmt.Errorf("chunk data too short: %d bytes", len(data))
	}

	// 解析 chunk item (header 48 bytes)
	length := binary.LittleEndian.Uint64(data[0:8])
	chunkType := binary.LittleEndian.Uint64(data[24:32])
	numStripes := binary.LittleEndian.Uint16(data[44:46]) // offset 44, not 48

	// 检查是否是 RAID 类型（简化：只支持 SINGLE/DUP）
	raidBits := chunkType & 0x1F8 // bits 3-8

	if raidBits != 0 && raidBits != (1<<5) { // 允许 DUP (1<<5)
		// 跳过复杂的 RAID 类型
		return nil
	}

	if numStripes < 1 {
		return fmt.Errorf("invalid num_stripes: %d", numStripes)
	}

	// 读取第一个 stripe (starts at offset 48)
	stripeOffset := 48
	if stripeOffset+32 > len(data) {
		return fmt.Errorf("stripe data out of bounds")
	}

	devID := binary.LittleEndian.Uint64(data[stripeOffset:])
	physicalOffset := binary.LittleEndian.Uint64(data[stripeOffset+8:])

	// 创建 chunk mapping
	mapping := &ChunkMapping{
		LogicalStart:  logicalOffset,
		LogicalLength: length,
		PhysicalStart: physicalOffset,
		DeviceID:      devID,
	}

	l.manager.AddMapping(mapping)

	return nil
}
