package ondisk

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Superblock defines the Btrfs superblock structure.
type Superblock struct {
	Checksum            [32]byte  // CRC32C checksum
	FSID                [16]byte  // Filesystem UUID
	Bytenr              uint64    // Physical address
	Flags               uint64    // Flags
	Magic               [8]byte   // Magic number "_BHRfS_M"
	Generation          uint64    // Generation
	Root                uint64    // Root tree logical address
	ChunkRoot           uint64    // Chunk tree logical address
	LogRoot             uint64    // Log tree logical address
	LogRootTransid      uint64    // Log root transaction ID
	TotalBytes          uint64    // Total bytes
	BytesUsed           uint64    // Bytes used
	RootDirObjectid     uint64    // Root directory objectid
	NumDevices          uint64    // Device count
	SectorSize          uint32    // Sector size
	NodeSize            uint32    // Node size
	LeafSize            uint32    // Leaf size (deprecated; equals NodeSize)
	StripeSize          uint32    // Stripe size
	SysChunkArraySize   uint32    // System chunk array size
	ChunkRootGeneration uint64    // Chunk root generation
	CompatFlags         uint64    // Compatible flags
	CompatRoFlags       uint64    // Read-only compatible flags
	IncompatFlags       uint64    // Incompatible flags
	CsumType            uint16    // Checksum type
	RootLevel           uint8     // Root tree level
	ChunkRootLevel      uint8     // Chunk root level
	LogRootLevel        uint8     // Log root level
	DevItem             DevItem   // Device info
	Label               [256]byte // Volume label
	CacheGeneration     uint64    // Cache generation
	UUIDTreeGeneration  uint64    // UUID tree generation
	MetadataUUID        [16]byte  // Metadata UUID

	// System chunk array (embedded chunk mapping)
	SysChunkArray [2048]byte
}

// DevItem represents a device item.
type DevItem struct {
	DevID       uint64   // Device ID
	TotalBytes  uint64   // Total device size
	BytesUsed   uint64   // Bytes used
	IOAlign     uint32   // IO alignment
	IOWidth     uint32   // IO width
	SectorSize  uint32   // Sector size
	Type        uint64   // Device type
	Generation  uint64   // Generation
	StartOffset uint64   // Start offset
	DevGroup    uint32   // Device group
	SeekSpeed   uint8    // Seek speed
	Bandwidth   uint8    // Bandwidth
	UUID        [16]byte // Device UUID
	FSID        [16]byte // Filesystem UUID
}

// Unmarshal parses a Superblock from a byte slice.
func (sb *Superblock) Unmarshal(data []byte) error {
	if len(data) < SuperblockSize {
		return fmt.Errorf("superblock data too short: got %d, need %d", len(data), SuperblockSize)
	}

	r := bytes.NewReader(data)

	// Read checksum.
	if err := binary.Read(r, binary.LittleEndian, &sb.Checksum); err != nil {
		return fmt.Errorf("failed to read checksum: %w", err)
	}

	// Read FSID.
	if err := binary.Read(r, binary.LittleEndian, &sb.FSID); err != nil {
		return fmt.Errorf("failed to read FSID: %w", err)
	}

	// Read basic fields.
	if err := binary.Read(r, binary.LittleEndian, &sb.Bytenr); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &sb.Flags); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &sb.Magic); err != nil {
		return err
	}

	// Validate magic number.
	if !bytes.Equal(sb.Magic[:], BtrfsMagic[:]) {
		return fmt.Errorf("invalid magic number: got %v, want %v", sb.Magic, BtrfsMagic)
	}

	// Read remaining fields.
	if err := binary.Read(r, binary.LittleEndian, &sb.Generation); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &sb.Root); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &sb.ChunkRoot); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &sb.LogRoot); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &sb.LogRootTransid); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &sb.TotalBytes); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &sb.BytesUsed); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &sb.RootDirObjectid); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &sb.NumDevices); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &sb.SectorSize); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &sb.NodeSize); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &sb.LeafSize); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &sb.StripeSize); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &sb.SysChunkArraySize); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &sb.ChunkRootGeneration); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &sb.CompatFlags); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &sb.CompatRoFlags); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &sb.IncompatFlags); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &sb.CsumType); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &sb.RootLevel); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &sb.ChunkRootLevel); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &sb.LogRootLevel); err != nil {
		return err
	}

	// Read DevItem.
	if err := sb.DevItem.Unmarshal(data[r.Size()-int64(r.Len()):]); err != nil {
		return fmt.Errorf("failed to read dev item: %w", err)
	}

	// Skip the already read DevItem size.
	devItemSize := 98 // DevItem struct size
	if _, err := r.Seek(int64(devItemSize), 1); err != nil {
		return err
	}

	// Read Label.
	if err := binary.Read(r, binary.LittleEndian, &sb.Label); err != nil {
		return err
	}

	// Jump from label directly to sys_chunk_array (offset 811).
	// Label ends at offset 555 (32+16+8*10+4*4+4+8*4+2+3+98+256).
	// sys_chunk_array starts at offset 811.
	// Skip 256 bytes (811 - 555 = 256).
	if _, err := r.Seek(256, 1); err != nil {
		return err
	}

	// Read system chunk array.
	if err := binary.Read(r, binary.LittleEndian, &sb.SysChunkArray); err != nil {
		return err
	}

	return nil
}

// Unmarshal parses a DevItem from a byte slice.
func (di *DevItem) Unmarshal(data []byte) error {
	r := bytes.NewReader(data)

	if err := binary.Read(r, binary.LittleEndian, &di.DevID); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &di.TotalBytes); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &di.BytesUsed); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &di.IOAlign); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &di.IOWidth); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &di.SectorSize); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &di.Type); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &di.Generation); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &di.StartOffset); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &di.DevGroup); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &di.SeekSpeed); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &di.Bandwidth); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &di.UUID); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &di.FSID); err != nil {
		return err
	}

	return nil
}

// GetLabel returns the volume label (trim trailing nulls).
func (sb *Superblock) GetLabel() string {
	// Find the first null byte.
	end := bytes.IndexByte(sb.Label[:], 0)
	if end == -1 {
		end = len(sb.Label)
	}
	return string(sb.Label[:end])
}

// IsValid validates whether the superblock is valid.
func (sb *Superblock) IsValid() bool {
	return bytes.Equal(sb.Magic[:], BtrfsMagic[:])
}
