package ondisk

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Superblock Btrfs 超级块结构
type Superblock struct {
	Checksum            [32]byte  // CRC32C 校验和
	FSID                [16]byte  // 文件系统 UUID
	Bytenr              uint64    // 物理地址
	Flags               uint64    // 标志位
	Magic               [8]byte   // 魔数 "_BHRfS_M"
	Generation          uint64    // 生成号
	Root                uint64    // Root tree 逻辑地址
	ChunkRoot           uint64    // Chunk tree 逻辑地址
	LogRoot             uint64    // Log tree 逻辑地址
	LogRootTransid      uint64    // Log root 事务 ID
	TotalBytes          uint64    // 总字节数
	BytesUsed           uint64    // 已使用字节数
	RootDirObjectid     uint64    // 根目录 objectid
	NumDevices          uint64    // 设备数量
	SectorSize          uint32    // 扇区大小
	NodeSize            uint32    // 节点大小
	LeafSize            uint32    // 叶节点大小 (已废弃，等于 NodeSize)
	StripeSize          uint32    // 条带大小
	SysChunkArraySize   uint32    // 系统 chunk 数组大小
	ChunkRootGeneration uint64    // Chunk root 生成号
	CompatFlags         uint64    // 兼容标志
	CompatRoFlags       uint64    // 只读兼容标志
	IncompatFlags       uint64    // 不兼容标志
	CsumType            uint16    // 校验和类型
	RootLevel           uint8     // Root tree 层级
	ChunkRootLevel      uint8     // Chunk root 层级
	LogRootLevel        uint8     // Log root 层级
	DevItem             DevItem   // 设备信息
	Label               [256]byte // 卷标签
	CacheGeneration     uint64    // 缓存生成号
	UUIDTreeGeneration  uint64    // UUID tree 生成号
	MetadataUUID        [16]byte  // 元数据 UUID

	// 系统 chunk 数组 (嵌入的 chunk 映射)
	SysChunkArray [2048]byte
}

// DevItem 设备项
type DevItem struct {
	DevID       uint64   // 设备 ID
	TotalBytes  uint64   // 设备总大小
	BytesUsed   uint64   // 已使用大小
	IOAlign     uint32   // IO 对齐
	IOWidth     uint32   // IO 宽度
	SectorSize  uint32   // 扇区大小
	Type        uint64   // 设备类型
	Generation  uint64   // 生成号
	StartOffset uint64   // 起始偏移
	DevGroup    uint32   // 设备组
	SeekSpeed   uint8    // 寻道速度
	Bandwidth   uint8    // 带宽
	UUID        [16]byte // 设备 UUID
	FSID        [16]byte // 文件系统 UUID
}

// Unmarshal 从字节数组解析 Superblock
func (sb *Superblock) Unmarshal(data []byte) error {
	if len(data) < SuperblockSize {
		return fmt.Errorf("superblock data too short: got %d, need %d", len(data), SuperblockSize)
	}

	r := bytes.NewReader(data)

	// 读取校验和
	if err := binary.Read(r, binary.LittleEndian, &sb.Checksum); err != nil {
		return fmt.Errorf("failed to read checksum: %w", err)
	}

	// 读取 FSID
	if err := binary.Read(r, binary.LittleEndian, &sb.FSID); err != nil {
		return fmt.Errorf("failed to read FSID: %w", err)
	}

	// 读取基本字段
	if err := binary.Read(r, binary.LittleEndian, &sb.Bytenr); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &sb.Flags); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &sb.Magic); err != nil {
		return err
	}

	// 验证魔数
	if !bytes.Equal(sb.Magic[:], BtrfsMagic[:]) {
		return fmt.Errorf("invalid magic number: got %v, want %v", sb.Magic, BtrfsMagic)
	}

	// 读取其他字段
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

	// 读取 DevItem
	if err := sb.DevItem.Unmarshal(data[r.Size()-int64(r.Len()):]); err != nil {
		return fmt.Errorf("failed to read dev item: %w", err)
	}

	// 跳过已读取的 DevItem 大小
	devItemSize := 98 // DevItem 结构体大小
	if _, err := r.Seek(int64(devItemSize), 1); err != nil {
		return err
	}

	// 读取 Label
	if err := binary.Read(r, binary.LittleEndian, &sb.Label); err != nil {
		return err
	}

	// 从 label 后直接跳到 sys_chunk_array (offset 811)
	// label 结束于 offset 555 (32+16+8*10+4*4+4+8*4+2+3+98+256)
	// sys_chunk_array 开始于 offset 811
	// 需要跳过 256 字节 (811 - 555 = 256)
	if _, err := r.Seek(256, 1); err != nil {
		return err
	}

	// 读取系统 chunk 数组
	if err := binary.Read(r, binary.LittleEndian, &sb.SysChunkArray); err != nil {
		return err
	}

	return nil
}

// Unmarshal 从字节数组解析 DevItem
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

// GetLabel 获取卷标签 (移除尾部的 null 字符)
func (sb *Superblock) GetLabel() string {
	// 查找第一个 null 字符
	end := bytes.IndexByte(sb.Label[:], 0)
	if end == -1 {
		end = len(sb.Label)
	}
	return string(sb.Label[:end])
}

// IsValid 验证 superblock 是否有效
func (sb *Superblock) IsValid() bool {
	return bytes.Equal(sb.Magic[:], BtrfsMagic[:])
}
