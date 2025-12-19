package chunk

import (
	"fmt"
)

// ChunkMapping Chunk 映射项
type ChunkMapping struct {
	LogicalStart  uint64
	LogicalLength uint64
	PhysicalStart uint64 // 简化版：只支持单个物理位置
	DeviceID      uint64
}

// PhysicalAddr 物理地址
type PhysicalAddr struct {
	DeviceID uint64
	Offset   uint64
}

// Contains 检查逻辑地址是否在此 chunk 范围内
func (c *ChunkMapping) Contains(logical uint64) bool {
	return logical >= c.LogicalStart && logical < c.LogicalStart+c.LogicalLength
}

// MapAddress 将逻辑地址映射到物理地址（简化版：SINGLE 类型）
func (c *ChunkMapping) MapAddress(logical uint64) (*PhysicalAddr, error) {
	if !c.Contains(logical) {
		return nil, fmt.Errorf("logical address 0x%x not in chunk range [0x%x, 0x%x)",
			logical, c.LogicalStart, c.LogicalStart+c.LogicalLength)
	}

	offsetInChunk := logical - c.LogicalStart

	return &PhysicalAddr{
		DeviceID: c.DeviceID,
		Offset:   c.PhysicalStart + offsetInChunk,
	}, nil
}
