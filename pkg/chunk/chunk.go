package chunk

import (
	"fmt"
)

// ChunkMapping represents a chunk mapping.
type ChunkMapping struct {
	LogicalStart  uint64
	LogicalLength uint64
	PhysicalStart uint64 // Simplified: supports a single physical location.
	DeviceID      uint64
}

// PhysicalAddr represents a physical address.
type PhysicalAddr struct {
	DeviceID uint64
	Offset   uint64
}

// Contains reports whether a logical address is in this chunk range.
func (c *ChunkMapping) Contains(logical uint64) bool {
	return logical >= c.LogicalStart && logical < c.LogicalStart+c.LogicalLength
}

// MapAddress maps a logical address to a physical address (simplified: SINGLE type).
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
