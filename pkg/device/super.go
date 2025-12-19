package device

import (
	"bytes"
	"fmt"

	"github.com/yourname/btrfs-read/pkg/errors"
	"github.com/yourname/btrfs-read/pkg/ondisk"
)

const (
	SuperblockOffset  int64 = 0x10000      // 64KB
	SuperblockBackup1 int64 = 0x4000000    // 64MB
	SuperblockBackup2 int64 = 0x4000000000 // 256GB
)

// SuperblockReader Superblock 读取器
type SuperblockReader struct {
	device BlockDevice
}

// NewSuperblockReader 创建读取器
func NewSuperblockReader(dev BlockDevice) *SuperblockReader {
	return &SuperblockReader{device: dev}
}

// ReadPrimary 读取主 Superblock
func (r *SuperblockReader) ReadPrimary() (*ondisk.Superblock, error) {
	return r.readAt(SuperblockOffset)
}

// ReadBackup 读取指定索引的备份 Superblock
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

// ReadLatest 读取最新的有效 Superblock（优先主，然后备份）
func (r *SuperblockReader) ReadLatest() (*ondisk.Superblock, error) {
	var candidates []*ondisk.Superblock
	var offsets = []int64{SuperblockOffset, SuperblockBackup1, SuperblockBackup2}

	// 尝试读取所有可能的 superblock
	for _, offset := range offsets {
		if offset >= r.device.Size() {
			continue
		}

		sb, err := r.readAt(offset)
		if err != nil {
			continue // 跳过无效的
		}

		if err := r.verify(sb); err != nil {
			continue // 跳过验证失败的
		}

		candidates = append(candidates, sb)
	}

	if len(candidates) == 0 {
		return nil, errors.ErrNoValidSuperblock
	}

	// 选择 generation 最大的
	latest := candidates[0]
	for _, sb := range candidates[1:] {
		if sb.Generation > latest.Generation {
			latest = sb
		}
	}

	return latest, nil
}

// readAt 从指定偏移读取 superblock
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

// verify 验证 superblock 的有效性
func (r *SuperblockReader) verify(sb *ondisk.Superblock) error {
	// 验证魔数
	if !bytes.Equal(sb.Magic[:], ondisk.BtrfsMagic[:]) {
		return errors.ErrInvalidMagic
	}

	// 验证基本字段
	if sb.SectorSize == 0 {
		return fmt.Errorf("invalid sector size: 0")
	}
	if sb.NodeSize == 0 {
		return fmt.Errorf("invalid node size: 0")
	}
	if sb.TotalBytes == 0 {
		return fmt.Errorf("invalid total bytes: 0")
	}

	// TODO: 验证 CRC32C 校验和（需要实现 checksum.go）

	return nil
}
