# Btrfs 文件系统读取服务 - 技术架构文档 (Golang 版本)

## 目录

1. [项目概述](#项目概述)
2. [技术选型](#技术选型)
3. [整体架构设计](#整体架构设计)
4. [核心模块设计](#核心模块设计)
5. [关键数据结构](#关键数据结构)
6. [核心流程详解](#核心流程详解)
7. [实现路线图](#实现路线图)
8. [性能优化策略](#性能优化策略)
9. [测试策略](#测试策略)
10. [参考资料](#参考资料)

---

## 项目概述

### 项目目标
开发一个只读的 Btrfs 文件系统读取服务，能够直接从块设备读取并解析 Btrfs 文件系统结构，提供文件和目录的访问能力。

### 核心特性
- ✅ 只读访问 Btrfs 文件系统
- ✅ 支持基本的文件读取和目录遍历
- ✅ 支持逻辑地址到物理地址映射 (Chunk Tree)
- ✅ 支持 B-Tree 索引遍历
- ✅ 支持数据压缩 (zlib, lzo, zstd)
- ✅ 支持校验和验证 (CRC32C)
- ✅ 支持多设备和基本 RAID (0/1)
- ❌ 不支持写操作
- ❌ 不支持扩展属性 (xattr)
- ❌ 不支持高级 RAID (5/6/10) - 可选实现

### 参考项目
本项目参考 [btrfs-fuse](https://github.com/adam900710/btrfs-fuse) 的架构设计，该项目是一个成熟的用户空间 Btrfs 只读实现。

---

## 技术选型

### 编程语言
**选择: Go (Golang) 1.21+**

理由:
- ✅ 开发效率高，语法简洁清晰
- ✅ 内置并发支持 (goroutine, channel)
- ✅ 丰富的标准库 (encoding/binary, io, os)
- ✅ 优秀的性能 (接近 C/C++)
- ✅ 跨平台编译简单
- ✅ 垃圾回收，内存管理简单
- ✅ 适合系统编程和文件系统开发

### 核心依赖库

```go
// go.mod
module github.com/WinBeyond/btrfs-read

go 1.21

require (
    // 压缩支持
    github.com/pierrec/lz4/v4 v4.1.18          // LZ4/LZO 压缩
    github.com/klauspost/compress v1.17.0      // Zstd 压缩 (高性能)
    // 注: zlib 使用标准库 compress/zlib
    
    // 校验和
    github.com/klauspost/crc32 v1.2.0          // CRC32C (SSE4.2/AVX2 优化)
    
    // 数据结构
    github.com/emirpasic/gods v1.18.1          // 红黑树等数据结构
    
    // FUSE 接口 (可选)
    github.com/hanwen/go-fuse/v2 v2.4.0        // Go-FUSE v2
    
    // 并发和同步
    golang.org/x/sync v0.5.0                   // errgroup, singleflight 等
)
```

### 构建工具
- **Go Modules**: Go 官方依赖管理
- **Makefile**: 自动化构建和测试

---

## 整体架构设计

### 分层架构

项目采用经典的分层架构，从底层到高层分为 5 层:

```
┌─────────────────────────────────────────────────────────┐
│              应用层 (Application Layer)                  │
│  - FUSE 接口 / CLI 工具                                  │
│  - 路径解析                                              │
│  - 用户 API                                              │
│  Package: cmd/, api/                                     │
└─────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────┐
│            文件系统层 (Filesystem Layer)                 │
│  - Inode 管理 (inode.go)                                │
│  - 目录遍历 (DIR_ITEM, DIR_INDEX)                       │
│  - 文件数据定位 (EXTENT_DATA)                           │
│  - 符号链接处理                                          │
│  Package: pkg/fs/                                        │
└─────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────┐
│              B-Tree 层 (B-Tree Layer)                    │
│  - B-Tree 遍历和查找 (btree.go)                         │
│  - Key 比较和搜索                                        │
│  - 节点解析 (Leaf/Internal Node)                        │
│  - Path 管理                                             │
│  Package: pkg/btree/                                     │
└─────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────┐
│            逻辑块层 (Logical Block Layer)                │
│  - Chunk Tree 解析 (chunk.go)                           │
│  - 逻辑地址 → 物理地址转换                               │
│  - RAID 处理 (raid.go)                                   │
│  - 多设备管理 (volumes.go)                              │
│  Package: pkg/chunk/                                     │
└─────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────┐
│            物理块层 (Physical Block Layer)               │
│  - 设备 I/O (device.go)                                  │
│  - Superblock 读取 (super.go)                           │
│  - 块缓存 (cache.go)                                     │
│  - 校验和验证 (checksum.go)                             │
│  - 解压缩 (compression.go)                              │
│  Package: pkg/device/                                    │
└─────────────────────────────────────────────────────────┘
```

参见架构图: [diagrams/architecture.md](../diagrams/architecture.md)

---

## 核心模块设计

### 项目目录结构

```
btrfs-read/
├── go.mod
├── go.sum
├── Makefile
├── README.md
├── ARCHITECTURE.md
│
├── cmd/
│   ├── btrfs-read/              # CLI 工具
│   │   └── main.go
│   └── btrfs-fuse/             # FUSE 挂载工具
│       └── main.go
│
├── pkg/
│   ├── ondisk/                 # 磁盘格式定义
│   │   ├── superblock.go       # Superblock 结构
│   │   ├── chunk.go            # Chunk 结构
│   │   ├── tree.go             # B-Tree 节点结构
│   │   ├── inode.go            # Inode 结构
│   │   ├── extent.go           # Extent 结构
│   │   └── constants.go        # 常量定义
│   │
│   ├── device/                 # 物理层
│   │   ├── device.go           # 块设备接口
│   │   ├── file_device.go      # 文件设备实现
│   │   ├── super.go            # Superblock 读取
│   │   ├── cache.go            # 块缓存
│   │   ├── checksum.go         # 校验和
│   │   └── compression.go      # 压缩/解压
│   │
│   ├── chunk/                  # 逻辑层
│   │   ├── chunk.go            # Chunk 映射
│   │   ├── manager.go          # Chunk 管理器
│   │   ├── volumes.go          # 多设备管理
│   │   └── raid.go             # RAID 支持
│   │
│   ├── btree/                  # B-Tree 层
│   │   ├── btree.go            # B-Tree 核心
│   │   ├── node.go             # 节点解析
│   │   ├── search.go           # 搜索算法
│   │   ├── key.go              # Key 操作
│   │   └── path.go             # Path 管理
│   │
│   ├── fs/                     # 文件系统层
│   │   ├── filesystem.go       # 文件系统主结构
│   │   ├── inode.go            # Inode 操作
│   │   ├── dir.go              # 目录操作
│   │   ├── file.go             # 文件读取
│   │   └── symlink.go          # 符号链接
│   │
│   ├── api/                    # 公共 API
│   │   └── api.go              # 外部接口定义
│   │
│   └── errors/                 # 错误定义
│       └── errors.go
│
├── internal/                   # 内部工具包
│   └── utils/
│       ├── hash.go             # 哈希函数
│       └── binary.go           # 二进制工具
│
├── diagrams/                   # 架构图 (Mermaid 格式)
│   ├── architecture.md       # 系统架构
│   ├── init-flow.md          # 初始化流程
│   ├── btree-search.md       # B-Tree 搜索
│   ├── file-read-flow.md     # 文件读取流程
│   ├── address-mapping.md    # 地址映射
│   └── README.md             # 图表说明
│
└── tests/                      # 测试
    ├── unit/                   # 单元测试
    ├── integration/            # 集成测试
    └── testdata/               # 测试数据
        └── test.img            # 测试镜像
```

---

## 核心模块详解

### 1. 物理层模块 (pkg/device/)

#### device.go - 块设备抽象

```go
package device

import "io"

// BlockDevice 块设备接口
type BlockDevice interface {
    // ReadAt 在指定偏移处读取数据
    ReadAt(p []byte, off int64) (n int, err error)
    
    // Size 返回设备大小
    Size() int64
    
    // DeviceID 返回设备唯一标识
    DeviceID() uint64
    
    // Close 关闭设备
    Close() error
}

// 确保实现了 io.ReaderAt 接口
var _ io.ReaderAt = (*FileDevice)(nil)
```

#### file_device.go - 文件设备实现

```go
package device

import (
    "os"
)

// FileDevice 文件后端块设备
type FileDevice struct {
    file     *os.File
    size     int64
    deviceID uint64
}

// NewFileDevice 创建文件设备
func NewFileDevice(path string) (*FileDevice, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    
    stat, err := file.Stat()
    if err != nil {
        file.Close()
        return nil, err
    }
    
    return &FileDevice{
        file:     file,
        size:     stat.Size(),
        deviceID: 0, // 从 superblock 读取
    }, nil
}

func (d *FileDevice) ReadAt(p []byte, off int64) (int, error) {
    return d.file.ReadAt(p, off)
}

func (d *FileDevice) Size() int64 {
    return d.size
}

func (d *FileDevice) DeviceID() uint64 {
    return d.deviceID
}

func (d *FileDevice) Close() error {
    return d.file.Close()
}
```

#### super.go - Superblock 管理

```go
package device

import (
    "bytes"
    "encoding/binary"
    "fmt"
    
    "github.com/WinBeyond/btrfs-read/pkg/ondisk"
)

const (
    SuperblockOffset   = 0x10000            // 64KB
    SuperblockBackup1  = 0x4000000          // 64MB
    SuperblockBackup2  = 0x4000000000       // 256GB
    SuperblockSize     = 4096
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

// ReadLatest 读取最新的有效 Superblock
func (r *SuperblockReader) ReadLatest() (*ondisk.Superblock, error) {
    // 尝试读取主 superblock
    primary, err := r.ReadPrimary()
    if err == nil && r.verify(primary) == nil {
        return primary, nil
    }
    
    // 尝试备份
    for _, offset := range []int64{SuperblockBackup1, SuperblockBackup2} {
        if offset > r.device.Size() {
            continue
        }
        sb, err := r.readAt(offset)
        if err == nil && r.verify(sb) == nil {
            return sb, nil
        }
    }
    
    return nil, fmt.Errorf("no valid superblock found")
}

func (r *SuperblockReader) readAt(offset int64) (*ondisk.Superblock, error) {
    buf := make([]byte, SuperblockSize)
    n, err := r.device.ReadAt(buf, offset)
    if err != nil || n != SuperblockSize {
        return nil, fmt.Errorf("failed to read superblock: %w", err)
    }
    
    sb := &ondisk.Superblock{}
    if err := sb.Unmarshal(buf); err != nil {
        return nil, err
    }
    
    return sb, nil
}

func (r *SuperblockReader) verify(sb *ondisk.Superblock) error {
    // 验证魔数
    if !bytes.Equal(sb.Magic[:], ondisk.BtrfsMagic[:]) {
        return fmt.Errorf("invalid magic number")
    }
    
    // 验证校验和
    // TODO: 实现 CRC32C 验证
    
    return nil
}
```

#### cache.go - 块缓存

```go
package device

import (
    "container/list"
    "sync"
)

// BlockCache LRU 块缓存
type BlockCache struct {
    mu         sync.RWMutex
    capacity   int
    cache      map[uint64]*list.Element
    lruList    *list.List
}

type cacheEntry struct {
    key  uint64
    data []byte
}

// NewBlockCache 创建缓存
func NewBlockCache(capacity int) *BlockCache {
    return &BlockCache{
        capacity: capacity,
        cache:    make(map[uint64]*list.Element),
        lruList:  list.New(),
    }
}

// Get 获取缓存项
func (c *BlockCache) Get(key uint64) ([]byte, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    if elem, ok := c.cache[key]; ok {
        c.lruList.MoveToFront(elem)
        return elem.Value.(*cacheEntry).data, true
    }
    return nil, false
}

// Put 添加缓存项
func (c *BlockCache) Put(key uint64, data []byte) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    if elem, ok := c.cache[key]; ok {
        c.lruList.MoveToFront(elem)
        elem.Value.(*cacheEntry).data = data
        return
    }
    
    entry := &cacheEntry{key: key, data: data}
    elem := c.lruList.PushFront(entry)
    c.cache[key] = elem
    
    if c.lruList.Len() > c.capacity {
        oldest := c.lruList.Back()
        if oldest != nil {
            c.lruList.Remove(oldest)
            delete(c.cache, oldest.Value.(*cacheEntry).key)
        }
    }
}

// Clear 清空缓存
func (c *BlockCache) Clear() {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    c.cache = make(map[uint64]*list.Element)
    c.lruList = list.New()
}
```

#### checksum.go - 校验和验证

```go
package device

import (
    "bytes"
    "crypto/sha256"
    "hash/crc32"
    
    crc32c "github.com/klauspost/crc32"
    "golang.org/x/crypto/blake2b"
)

// ChecksumType 校验和类型
type ChecksumType uint8

const (
    ChecksumCRC32C ChecksumType = iota
    ChecksumXXHash
    ChecksumSHA256
    ChecksumBlake2
)

// ChecksumVerifier 校验和验证器
type ChecksumVerifier struct{}

// Compute 计算校验和
func (v *ChecksumVerifier) Compute(data []byte, csumType ChecksumType) ([]byte, error) {
    switch csumType {
    case ChecksumCRC32C:
        // 使用 Castagnoli 多项式 (SSE4.2 优化)
        table := crc32c.MakeTable(crc32c.Castagnoli)
        sum := crc32c.Checksum(data, table)
        result := make([]byte, 4)
        binary.LittleEndian.PutUint32(result, sum)
        return result, nil
        
    case ChecksumSHA256:
        hash := sha256.Sum256(data)
        return hash[:], nil
        
    case ChecksumBlake2:
        hash := blake2b.Sum256(data)
        return hash[:], nil
        
    default:
        return nil, fmt.Errorf("unsupported checksum type: %d", csumType)
    }
}

// Verify 验证校验和
func (v *ChecksumVerifier) Verify(data []byte, expected []byte, csumType ChecksumType) error {
    computed, err := v.Compute(data, csumType)
    if err != nil {
        return err
    }
    
    if !bytes.Equal(computed, expected) {
        return fmt.Errorf("checksum mismatch")
    }
    
    return nil
}
```

#### compression.go - 压缩支持

```go
package device

import (
    "bytes"
    "compress/zlib"
    "fmt"
    "io"
    
    "github.com/klauspost/compress/zstd"
    "github.com/pierrec/lz4/v4"
)

// CompressionType 压缩类型
type CompressionType uint8

const (
    CompressionNone CompressionType = iota
    CompressionZlib
    CompressionLZO
    CompressionZstd
    CompressionLZ4
)

// Decompressor 解压缩器
type Decompressor struct{}

// Decompress 解压缩数据
func (d *Decompressor) Decompress(compressed []byte, compType CompressionType, decompressedSize int) ([]byte, error) {
    switch compType {
    case CompressionNone:
        return compressed, nil
        
    case CompressionZlib:
        return d.decompressZlib(compressed)
        
    case CompressionLZO, CompressionLZ4:
        return d.decompressLZ4(compressed, decompressedSize)
        
    case CompressionZstd:
        return d.decompressZstd(compressed)
        
    default:
        return nil, fmt.Errorf("unsupported compression type: %d", compType)
    }
}

func (d *Decompressor) decompressZlib(compressed []byte) ([]byte, error) {
    reader, err := zlib.NewReader(bytes.NewReader(compressed))
    if err != nil {
        return nil, err
    }
    defer reader.Close()
    
    var buf bytes.Buffer
    if _, err := io.Copy(&buf, reader); err != nil {
        return nil, err
    }
    
    return buf.Bytes(), nil
}

func (d *Decompressor) decompressLZ4(compressed []byte, size int) ([]byte, error) {
    reader := lz4.NewReader(bytes.NewReader(compressed))
    decompressed := make([]byte, size)
    
    n, err := io.ReadFull(reader, decompressed)
    if err != nil && err != io.ErrUnexpectedEOF {
        return nil, err
    }
    
    return decompressed[:n], nil
}

func (d *Decompressor) decompressZstd(compressed []byte) ([]byte, error) {
    decoder, err := zstd.NewReader(nil)
    if err != nil {
        return nil, err
    }
    defer decoder.Close()
    
    return decoder.DecodeAll(compressed, nil)
}
```

### 2. 逻辑层模块 (pkg/chunk/)

#### chunk.go - Chunk 映射

```go
package chunk

import (
    "fmt"
)

// RaidType RAID 类型
type RaidType uint64

const (
    RaidSingle RaidType = 0
    RaidRaid0  RaidType = 1 << 3
    RaidRaid1  RaidType = 1 << 4
    RaidDup    RaidType = 1 << 5
    RaidRaid10 RaidType = 1 << 6
    RaidRaid5  RaidType = 1 << 7
    RaidRaid6  RaidType = 1 << 8
)

// ChunkMapping Chunk 映射项
type ChunkMapping struct {
    LogicalStart  uint64
    LogicalLength uint64
    StripeLength  uint64
    NumStripes    uint16
    RaidType      RaidType
    Stripes       []Stripe
}

// Stripe 条带信息
type Stripe struct {
    DeviceID uint64
    Offset   uint64
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

// MapAddress 将逻辑地址映射到物理地址
func (c *ChunkMapping) MapAddress(logical uint64) ([]PhysicalAddr, error) {
    if !c.Contains(logical) {
        return nil, fmt.Errorf("logical address 0x%x not in chunk range", logical)
    }
    
    offsetInChunk := logical - c.LogicalStart
    
    switch c.RaidType {
    case RaidSingle, RaidDup:
        return []PhysicalAddr{
            {DeviceID: c.Stripes[0].DeviceID, Offset: c.Stripes[0].Offset + offsetInChunk},
        }, nil
        
    case RaidRaid0:
        return c.mapRaid0(offsetInChunk)
        
    case RaidRaid1:
        return c.mapRaid1(offsetInChunk)
        
    default:
        return nil, fmt.Errorf("unsupported RAID type: 0x%x", c.RaidType)
    }
}

func (c *ChunkMapping) mapRaid0(offsetInChunk uint64) ([]PhysicalAddr, error) {
    stripeNr := offsetInChunk / c.StripeLength
    stripeIndex := int(stripeNr % uint64(c.NumStripes))
    stripeOffset := (stripeNr / uint64(c.NumStripes)) * c.StripeLength
    offsetInStripe := offsetInChunk % c.StripeLength
    
    stripe := c.Stripes[stripeIndex]
    return []PhysicalAddr{
        {DeviceID: stripe.DeviceID, Offset: stripe.Offset + stripeOffset + offsetInStripe},
    }, nil
}

func (c *ChunkMapping) mapRaid1(offsetInChunk uint64) ([]PhysicalAddr, error) {
    // RAID1: 返回所有镜像的地址
    addrs := make([]PhysicalAddr, len(c.Stripes))
    for i, stripe := range c.Stripes {
        addrs[i] = PhysicalAddr{
            DeviceID: stripe.DeviceID,
            Offset:   stripe.Offset + offsetInChunk,
        }
    }
    return addrs, nil
}
```

#### manager.go - Chunk 管理器

```go
package chunk

import (
    "fmt"
    "sync"
    
    rbtree "github.com/emirpasic/gods/trees/redblacktree"
)

// Manager Chunk 管理器
type Manager struct {
    mu       sync.RWMutex
    mappings *rbtree.Tree // key: LogicalStart, value: *ChunkMapping
}

// NewManager 创建管理器
func NewManager() *Manager {
    return &Manager{
        mappings: rbtree.NewWithIntComparator(),
    }
}

// AddMapping 添加 Chunk 映射
func (m *Manager) AddMapping(mapping *ChunkMapping) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    m.mappings.Put(mapping.LogicalStart, mapping)
}

// LogicalToPhysical 逻辑地址转物理地址
func (m *Manager) LogicalToPhysical(logical uint64) (*PhysicalAddr, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    // 查找包含该逻辑地址的 chunk
    mapping := m.findMapping(logical)
    if mapping == nil {
        return nil, fmt.Errorf("no chunk mapping found for logical 0x%x", logical)
    }
    
    // 映射地址
    addrs, err := mapping.MapAddress(logical)
    if err != nil {
        return nil, err
    }
    
    // 返回第一个地址 (对于 RAID1，可以选择任意镜像)
    return &addrs[0], nil
}

func (m *Manager) findMapping(logical uint64) *ChunkMapping {
    // 查找 <= logical 的最大 key
    var found *ChunkMapping
    
    it := m.mappings.Iterator()
    for it.Next() {
        mapping := it.Value().(*ChunkMapping)
        if mapping.Contains(logical) {
            found = mapping
            break
        }
    }
    
    return found
}
```

### 3. B-Tree 层模块 (pkg/btree/)

#### key.go - Key 操作

```go
package btree

import (
    "encoding/binary"
)

// Key 类型常量
const (
    KeyTypeInodeItem   uint8 = 1
    KeyTypeInodeRef    uint8 = 12
    KeyTypeXattrItem   uint8 = 24
    KeyTypeExtentData  uint8 = 108
    KeyTypeDirItem     uint8 = 84
    KeyTypeDirIndex    uint8 = 96
    KeyTypeChunkItem   uint8 = 228
)

// Key Btrfs B-Tree Key
type Key struct {
    ObjectID uint64 // inode 号或其他对象 ID
    Type     uint8  // key 类型
    Offset   uint64 // 偏移或哈希值
}

// Compare 比较两个 Key
// 返回: -1 (this < other), 0 (==), 1 (this > other)
func (k *Key) Compare(other *Key) int {
    if k.ObjectID < other.ObjectID {
        return -1
    } else if k.ObjectID > other.ObjectID {
        return 1
    }
    
    if k.Type < other.Type {
        return -1
    } else if k.Type > other.Type {
        return 1
    }
    
    if k.Offset < other.Offset {
        return -1
    } else if k.Offset > other.Offset {
        return 1
    }
    
    return 0
}

// Unmarshal 从字节数组解析
func (k *Key) Unmarshal(data []byte) error {
    if len(data) < 17 {
        return fmt.Errorf("key data too short")
    }
    
    k.ObjectID = binary.LittleEndian.Uint64(data[0:8])
    k.Type = data[8]
    k.Offset = binary.LittleEndian.Uint64(data[9:17])
    
    return nil
}
```

#### node.go - 节点解析

```go
package btree

import (
    "encoding/binary"
    "fmt"
)

const (
    HeaderSize = 101
)

// Header B-Tree 节点头
type Header struct {
    Checksum [32]byte
    FSID     [16]byte
    Bytenr   uint64
    Flags    uint64
    ChunkUUID [16]byte
    Generation uint64
    Owner    uint64
    NrItems  uint32
    Level    uint8
}

// Node B-Tree 节点
type Node struct {
    Header  *Header
    Keys    []*Key
    Ptrs    []uint64      // 内部节点: 子节点指针
    Items   []*Item       // 叶节点: item 数据
}

// Item 叶节点中的 item
type Item struct {
    Key    *Key
    Offset uint32 // 数据在节点中的偏移
    Size   uint32 // 数据大小
    Data   []byte // 实际数据
}

// IsLeaf 是否是叶节点
func (h *Header) IsLeaf() bool {
    return h.Level == 0
}

// UnmarshalNode 解析节点
func UnmarshalNode(data []byte, nodeSize uint32) (*Node, error) {
    if len(data) < int(nodeSize) {
        return nil, fmt.Errorf("node data too short")
    }
    
    // 解析 header
    header := &Header{}
    if err := header.Unmarshal(data[:HeaderSize]); err != nil {
        return nil, err
    }
    
    node := &Node{Header: header}
    
    if header.IsLeaf() {
        return unmarshalLeafNode(node, data)
    }
    return unmarshalInternalNode(node, data)
}

func (h *Header) Unmarshal(data []byte) error {
    copy(h.Checksum[:], data[0:32])
    copy(h.FSID[:], data[32:48])
    h.Bytenr = binary.LittleEndian.Uint64(data[48:56])
    h.Flags = binary.LittleEndian.Uint64(data[56:64])
    copy(h.ChunkUUID[:], data[64:80])
    h.Generation = binary.LittleEndian.Uint64(data[80:88])
    h.Owner = binary.LittleEndian.Uint64(data[88:96])
    h.NrItems = binary.LittleEndian.Uint32(data[96:100])
    h.Level = data[100]
    return nil
}

func unmarshalLeafNode(node *Node, data []byte) (*Node, error) {
    node.Items = make([]*Item, node.Header.NrItems)
    
    offset := HeaderSize
    for i := uint32(0); i < node.Header.NrItems; i++ {
        item := &Item{}
        
        // 解析 key (17 bytes)
        key := &Key{}
        if err := key.Unmarshal(data[offset : offset+17]); err != nil {
            return nil, err
        }
        item.Key = key
        offset += 17
        
        // 解析 offset 和 size
        item.Offset = binary.LittleEndian.Uint32(data[offset : offset+4])
        item.Size = binary.LittleEndian.Uint32(data[offset+4 : offset+8])
        offset += 8
        
        node.Items[i] = item
    }
    
    // 读取实际数据
    for _, item := range node.Items {
        dataOffset := HeaderSize + int(item.Offset)
        item.Data = data[dataOffset : dataOffset+int(item.Size)]
    }
    
    return node, nil
}

func unmarshalInternalNode(node *Node, data []byte) (*Node, error) {
    node.Keys = make([]*Key, node.Header.NrItems)
    node.Ptrs = make([]uint64, node.Header.NrItems)
    
    offset := HeaderSize
    for i := uint32(0); i < node.Header.NrItems; i++ {
        // 解析 key (17 bytes)
        key := &Key{}
        if err := key.Unmarshal(data[offset : offset+17]); err != nil {
            return nil, err
        }
        node.Keys[i] = key
        offset += 17
        
        // 解析 block pointer (8 bytes)
        node.Ptrs[i] = binary.LittleEndian.Uint64(data[offset : offset+8])
        offset += 8
        
        // 跳过 generation (8 bytes)
        offset += 8
    }
    
    return node, nil
}
```

#### search.go - 搜索算法

```go
package btree

import (
    "fmt"
)

// Path B-Tree 搜索路径
type Path struct {
    Nodes []*Node
    Slots []int
}

// Searcher B-Tree 搜索器
type Searcher struct {
    reader NodeReader
}

// NodeReader 节点读取接口
type NodeReader interface {
    ReadNode(logical uint64) (*Node, error)
}

// NewSearcher 创建搜索器
func NewSearcher(reader NodeReader) *Searcher {
    return &Searcher{reader: reader}
}

// Search 搜索指定的 key
func (s *Searcher) Search(rootAddr uint64, targetKey *Key) (*Path, error) {
    path := &Path{
        Nodes: make([]*Node, 0),
        Slots: make([]int, 0),
    }
    
    currentAddr := rootAddr
    
    for {
        // 读取当前节点
        node, err := s.reader.ReadNode(currentAddr)
        if err != nil {
            return nil, fmt.Errorf("failed to read node: %w", err)
        }
        
        // 二分查找
        slot := s.binarySearch(node, targetKey)
        
        path.Nodes = append(path.Nodes, node)
        path.Slots = append(path.Slots, slot)
        
        // 如果是叶节点，搜索完成
        if node.Header.IsLeaf() {
            return path, nil
        }
        
        // 内部节点，继续向下
        if slot >= len(node.Ptrs) {
            return nil, fmt.Errorf("invalid slot %d", slot)
        }
        currentAddr = node.Ptrs[slot]
    }
}

// binarySearch 二分查找
func (s *Searcher) binarySearch(node *Node, targetKey *Key) int {
    if node.Header.IsLeaf() {
        return s.binarySearchLeaf(node, targetKey)
    }
    return s.binarySearchInternal(node, targetKey)
}

func (s *Searcher) binarySearchLeaf(node *Node, targetKey *Key) int {
    left, right := 0, len(node.Items)
    
    for left < right {
        mid := (left + right) / 2
        cmp := node.Items[mid].Key.Compare(targetKey)
        
        if cmp < 0 {
            left = mid + 1
        } else {
            right = mid
        }
    }
    
    return left
}

func (s *Searcher) binarySearchInternal(node *Node, targetKey *Key) int {
    left, right := 0, len(node.Keys)
    
    for left < right {
        mid := (left + right) / 2
        cmp := node.Keys[mid].Compare(targetKey)
        
        if cmp < 0 {
            left = mid + 1
        } else {
            right = mid
        }
    }
    
    // 内部节点：如果没有精确匹配，使用前一个指针
    if left > 0 && (left >= len(node.Keys) || node.Keys[left].Compare(targetKey) > 0) {
        left--
    }
    
    return left
}
```

### 4. 文件系统层 (pkg/fs/)

由于篇幅限制，这里展示核心接口:

#### filesystem.go - 主结构

```go
package fs

import (
    "github.com/WinBeyond/btrfs-read/pkg/btree"
    "github.com/WinBeyond/btrfs-read/pkg/chunk"
    "github.com/WinBeyond/btrfs-read/pkg/device"
    "github.com/WinBeyond/btrfs-read/pkg/ondisk"
)

// FileSystem Btrfs 文件系统
type FileSystem struct {
    superblock    *ondisk.Superblock
    volumeManager *VolumeManager
    chunkManager  *chunk.Manager
    cache         *device.BlockCache
    
    rootTreeRoot uint64
    fsTreeRoot   uint64
    
    btreeSearcher *btree.Searcher
}

// Open 打开文件系统
func Open(devicePaths []string) (*FileSystem, error) {
    // 1. 扫描设备
    // 2. 读取 superblock
    // 3. 初始化 chunk tree
    // 4. 定位 fs tree
    // 5. 创建缓存
    
    // TODO: 实现
    return nil, nil
}

// ReadFile 读取文件
func (fs *FileSystem) ReadFile(path string) ([]byte, error) {
    // 1. 路径解析 -> inode
    // 2. 读取 extent
    // 3. 读取数据
    // 4. 解压缩
    
    // TODO: 实现
    return nil, nil
}

// ReadDir 读取目录
func (fs *FileSystem) ReadDir(path string) ([]*DirEntry, error) {
    // TODO: 实现
    return nil, nil
}

// ReadLink 读取符号链接
func (fs *FileSystem) ReadLink(path string) (string, error) {
    // TODO: 实现
    return "", nil
}

// DirEntry 目录项
type DirEntry struct {
    Name     string
    Inode    uint64
    FileType FileType
}

// FileType 文件类型
type FileType uint8

const (
    FileTypeRegular FileType = iota
    FileTypeDirectory
    FileTypeSymlink
)
```

---

## 实现路线图

### Phase 0: 项目初始化 (1 天)

- [ ] 创建 Go 项目结构
- [ ] 初始化 go.mod，配置依赖
- [ ] 设置 Makefile
- [ ] 编写基础错误类型 (pkg/errors/)
- [ ] 定义磁盘格式常量 (pkg/ondisk/constants.go)

### Phase 1: 物理层实现 (3-5 天)

**目标**: 能够读取并解析 Superblock

- [ ] 实现 BlockDevice 接口 (pkg/device/device.go)
  - [ ] FileDevice 实现
- [ ] 实现 Superblock 解析 (pkg/device/super.go)
  - [ ] 读取主 Superblock
  - [ ] 验证魔数
  - [ ] CRC32C 校验和验证
- [ ] 实现 BlockCache (pkg/device/cache.go)

**验收标准:**
```go
func TestReadSuperblock(t *testing.T) {
    device, _ := device.NewFileDevice("test.btrfs")
    reader := device.NewSuperblockReader(device)
    sb, err := reader.ReadLatest()
    assert.NoError(t, err)
    assert.Equal(t, ondisk.BtrfsMagic, sb.Magic)
}
```

### Phase 2-7: 参见完整文档...

(其余阶段与 Rust 版本类似，主要是实现语言不同)

---

## 性能优化策略

### 1. Goroutine 并发

```go
// 并行读取多个文件
func (fs *FileSystem) ReadFilesParallel(paths []string) ([][]byte, error) {
    results := make([][]byte, len(paths))
    errChan := make(chan error, len(paths))
    
    var wg sync.WaitGroup
    for i, path := range paths {
        wg.Add(1)
        go func(index int, p string) {
            defer wg.Done()
            data, err := fs.ReadFile(p)
            if err != nil {
                errChan <- err
                return
            }
            results[index] = data
        }(i, path)
    }
    
    wg.Wait()
    close(errChan)
    
    if len(errChan) > 0 {
        return nil, <-errChan
    }
    
    return results, nil
}
```

### 2. 内存池

```go
var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 16384) // nodesize
    },
}

func (fs *FileSystem) readNodeOptimized(addr uint64) (*btree.Node, error) {
    buf := bufferPool.Get().([]byte)
    defer bufferPool.Put(buf)
    
    // 使用 buf 读取节点
    // ...
}
```

### 3. singleflight 优化

避免重复请求:
```go
import "golang.org/x/sync/singleflight"

type FileSystem struct {
    // ...
    sf singleflight.Group
}

func (fs *FileSystem) ReadNode(addr uint64) (*btree.Node, error) {
    key := fmt.Sprintf("node:%d", addr)
    
    v, err, _ := fs.sf.Do(key, func() (interface{}, error) {
        return fs.readNodeInternal(addr)
    })
    
    if err != nil {
        return nil, err
    }
    return v.(*btree.Node), nil
}
```

---

## 测试策略

### 单元测试

```go
// pkg/btree/key_test.go
func TestKeyCompare(t *testing.T) {
    k1 := &Key{ObjectID: 1, Type: 1, Offset: 0}
    k2 := &Key{ObjectID: 1, Type: 1, Offset: 1}
    
    assert.Equal(t, -1, k1.Compare(k2))
    assert.Equal(t, 1, k2.Compare(k1))
    assert.Equal(t, 0, k1.Compare(k1))
}
```

### 集成测试

```go
// tests/integration/read_test.go
func TestReadFile(t *testing.T) {
    fs, err := Open([]string{"testdata/test.img"})
    require.NoError(t, err)
    defer fs.Close()
    
    data, err := fs.ReadFile("/test.txt")
    require.NoError(t, err)
    assert.Equal(t, []byte("Hello Btrfs\n"), data)
}
```

### Benchmark

```go
func BenchmarkReadFile(b *testing.B) {
    fs, _ := Open([]string{"testdata/test.img"})
    defer fs.Close()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        fs.ReadFile("/test.txt")
    }
}
```

---

**文档版本**: v2.0 (Golang)  
**创建日期**: 2025-12-18  
**最后更新**: 2025-12-18  
**语言**: Go 1.21+
