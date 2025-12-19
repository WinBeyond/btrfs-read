package device

import (
	"bytes"
	"fmt"

	"github.com/WinBeyond/btrfs-read/pkg/errors"
	"github.com/WinBeyond/btrfs-read/pkg/ondisk"
)

const (
	SuperblockOffset  int64 = 0x10000      // 64KB
	SuperblockBackup1 int64 = 0x4000000    // 64MB
	SuperblockBackup2 int64 = 0x4000000000 // 256GB
)

// SuperblockReader reads superblocks.
type SuperblockReader struct {
	device BlockDevice
}

// NewSuperblockReader creates a reader.
func NewSuperblockReader(dev BlockDevice) *SuperblockReader {
	return &SuperblockReader{device: dev}
}

// ReadPrimary reads the primary superblock.
func (r *SuperblockReader) ReadPrimary() (*ondisk.Superblock, error) {
	return r.readAt(SuperblockOffset)
}

// ReadBackup reads a backup superblock at the given index.
// index: 0 = primary, 1 = backup1 (64MB), 2 = backup2 (256GB)
func (r *SuperblockReader) ReadBackup(index int) (*ondisk.Superblock, error) {
	var offset int64
	switch index {
	case 0:
		offset = SuperblockOffset
	case 1:
		offset = SuperblockBackup1
	case 2:
		offset = SuperblockBackup2
	default:
		return nil, fmt.Errorf("invalid backup index: %d", index)
	}

	if offset >= r.device.Size() {
		return nil, fmt.Errorf("backup offset %d beyond device size %d", offset, r.device.Size())
	}

	return r.readAt(offset)
}

// ReadLatest reads the newest valid superblock (primary then backups).
func (r *SuperblockReader) ReadLatest() (*ondisk.Superblock, error) {
	var candidates []*ondisk.Superblock
	var offsets = []int64{SuperblockOffset, SuperblockBackup1, SuperblockBackup2}

	// Try reading all possible superblocks.
	for _, offset := range offsets {
		if offset >= r.device.Size() {
			continue
		}

		sb, err := r.readAt(offset)
		if err != nil {
			continue // Skip invalid.
		}

		if err := r.verify(sb); err != nil {
			continue // Skip failed verification.
		}

		candidates = append(candidates, sb)
	}

	if len(candidates) == 0 {
		return nil, errors.ErrNoValidSuperblock
	}

	// Select the largest generation.
	latest := candidates[0]
	for _, sb := range candidates[1:] {
		if sb.Generation > latest.Generation {
			latest = sb
		}
	}

	return latest, nil
}

// readAt reads a superblock at the specified offset.
func (r *SuperblockReader) readAt(offset int64) (*ondisk.Superblock, error) {
	buf := make([]byte, ondisk.SuperblockSize)
	n, err := r.device.ReadAt(buf, offset)
	if err != nil {
		return nil, errors.Wrap("SuperblockReader.readAt", err)
	}
	if n != ondisk.SuperblockSize {
		return nil, errors.Wrap("SuperblockReader.readAt",
			fmt.Errorf("read %d bytes, expected %d", n, ondisk.SuperblockSize))
	}

	sb := &ondisk.Superblock{}
	if err := sb.Unmarshal(buf); err != nil {
		return nil, errors.Wrap("SuperblockReader.Unmarshal", err)
	}

	return sb, nil
}

// verify validates the superblock.
func (r *SuperblockReader) verify(sb *ondisk.Superblock) error {
	// Validate magic number.
	if !bytes.Equal(sb.Magic[:], ondisk.BtrfsMagic[:]) {
		return errors.ErrInvalidMagic
	}

	// Validate basic fields.
	if sb.SectorSize == 0 {
		return fmt.Errorf("invalid sector size: 0")
	}
	if sb.NodeSize == 0 {
		return fmt.Errorf("invalid node size: 0")
	}
	if sb.TotalBytes == 0 {
		return fmt.Errorf("invalid total bytes: 0")
	}

	// TODO: Validate CRC32C checksum (requires checksum.go).

	return nil
}
