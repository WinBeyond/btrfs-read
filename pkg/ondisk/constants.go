package ondisk

// Btrfs magic number.
var BtrfsMagic = [8]byte{'_', 'B', 'H', 'R', 'f', 'S', '_', 'M'}

// Superblock-related constants.
const (
	SuperblockOffset  int64 = 0x10000      // 64KB
	SuperblockBackup1 int64 = 0x4000000    // 64MB
	SuperblockBackup2 int64 = 0x4000000000 // 256GB
	SuperblockSize    int   = 4096
	ChecksumSize      int   = 32
	UUIDSize          int   = 16
	FsidSize          int   = 16
)

// B-Tree-related constants.
const (
	HeaderSize        = 101
	DefaultNodeSize   = 16384 // 16KB
	DefaultLeafSize   = 16384
	DefaultSectorSize = 4096
)

// Key types.
const (
	KeyTypeInodeItem       uint8 = 1
	KeyTypeInodeRef        uint8 = 12
	KeyTypeDirLog          uint8 = 60
	KeyTypeDirLogIndex     uint8 = 72
	KeyTypeXattrItem       uint8 = 24
	KeyTypeExtentData      uint8 = 108
	KeyTypeExtentCsum      uint8 = 128
	KeyTypeRootItem        uint8 = 132
	KeyTypeRootBackref     uint8 = 144
	KeyTypeRootRef         uint8 = 156
	KeyTypeExtentItem      uint8 = 168
	KeyTypeMetadataItem    uint8 = 169
	KeyTypeTreeBlockRef    uint8 = 176
	KeyTypeExtentDataRef   uint8 = 178
	KeyTypeExtentRefV0     uint8 = 180
	KeyTypeSharedBlockRef  uint8 = 182
	KeyTypeSharedDataRef   uint8 = 184
	KeyTypeBlockGroupItem  uint8 = 192
	KeyTypeFreeSpaceInfo   uint8 = 198
	KeyTypeFreeSpaceExtent uint8 = 199
	KeyTypeFreeSpaceBitmap uint8 = 200
	KeyTypeDevExtent       uint8 = 204
	KeyTypeDevItem         uint8 = 216
	KeyTypeChunkItem       uint8 = 228
	KeyTypeDirItem         uint8 = 84
	KeyTypeDirIndex        uint8 = 96
)

// Object ID
const (
	RootTreeObjectid      uint64 = 1
	ExtentTreeObjectid    uint64 = 2
	ChunkTreeObjectid     uint64 = 3
	DevTreeObjectid       uint64 = 4
	FsTreeObjectid        uint64 = 5
	RootTreeDirObjectid   uint64 = 6
	CsumTreeObjectid      uint64 = 7
	QuotaTreeObjectid     uint64 = 8
	UUIDTreeObjectid      uint64 = 9
	FreeSpaceTreeObjectid uint64 = 10
	DevStatsObjectid      uint64 = 0
	BalanceObjectid       uint64 = 0xFFFFFFFFFFFFFFF4 // -4 in uint64
	OrphanObjectid        uint64 = 0xFFFFFFFFFFFFFFFB // -5
	TreeLogObjectid       uint64 = 0xFFFFFFFFFFFFFFFA // -6
	TreeLogFixupObjectid  uint64 = 0xFFFFFFFFFFFFFFF9 // -7
	TreeRelocObjectid     uint64 = 0xFFFFFFFFFFFFFFF8 // -8
	DataRelocTreeObjectid uint64 = 0xFFFFFFFFFFFFFFF7 // -9
	ExtentCsumObjectid    uint64 = 0xFFFFFFFFFFFFFFF6 // -10
	FreeSpaceObjectid     uint64 = 0xFFFFFFFFFFFFFFF5 // -11
	FreeInoObjectid       uint64 = 0xFFFFFFFFFFFFFFF4 // -12
	MultipleObjectids     uint64 = 0xFFFFFFFFFFFFFF01 // -255
	FirstFreeObjectid     uint64 = 256
	LastFreeObjectid      uint64 = 0xFFFFFFFFFFFFFF00 // -256
)

// File types.
const (
	FtUnknown uint8 = 0
	FtRegFile uint8 = 1
	FtDir     uint8 = 2
	FtChrdev  uint8 = 3
	FtBlkdev  uint8 = 4
	FtFifo    uint8 = 5
	FtSock    uint8 = 6
	FtSymlink uint8 = 7
	FtXattr   uint8 = 8
	FtMax     uint8 = 9
)

// File extent types.
const (
	FileExtentInline   uint8 = 0
	FileExtentReg      uint8 = 1
	FileExtentPrealloc uint8 = 2
)

// Compression types.
const (
	CompressNone uint8 = 0
	CompressZlib uint8 = 1
	CompressLZO  uint8 = 2
	CompressZstd uint8 = 3
	CompressLZ4  uint8 = 4 // Not supported by the btrfs kernel, but used by some implementations.
)

// Checksum types.
const (
	CsumTypeCRC32C uint16 = 0
	CsumTypeXXHash uint16 = 1
	CsumTypeSHA256 uint16 = 2
	CsumTypeBlake2 uint16 = 3
)

// RAID type flags.
const (
	BlockGroupData     uint64 = 1 << 0
	BlockGroupSystem   uint64 = 1 << 1
	BlockGroupMetadata uint64 = 1 << 2
	BlockGroupRaid0    uint64 = 1 << 3
	BlockGroupRaid1    uint64 = 1 << 4
	BlockGroupDup      uint64 = 1 << 5
	BlockGroupRaid10   uint64 = 1 << 6
	BlockGroupRaid5    uint64 = 1 << 7
	BlockGroupRaid6    uint64 = 1 << 8
	BlockGroupRaid1C3  uint64 = 1 << 9
	BlockGroupRaid1C4  uint64 = 1 << 10
)

// Inode flags.
const (
	InodeNodatasum  uint64 = 1 << 0  // Do not calculate data checksums.
	InodeNodatacow  uint64 = 1 << 1  // Do not use COW.
	InodeReadonly   uint64 = 1 << 2  // Read-only.
	InodeNocompress uint64 = 1 << 3  // Do not compress.
	InodePrealloc   uint64 = 1 << 4  // Preallocate.
	InodeSync       uint64 = 1 << 5  // Synchronous writes.
	InodeImmutable  uint64 = 1 << 6  // Immutable.
	InodeAppend     uint64 = 1 << 7  // Append-only.
	InodeNodump     uint64 = 1 << 8  // Exclude from dumps.
	InodeNoatime    uint64 = 1 << 9  // Do not update access time.
	InodeDirsync    uint64 = 1 << 10 // Directory sync.
	InodeCompress   uint64 = 1 << 11 // Compress.
)

// Root flags.
const (
	RootSubvolReadonly uint64 = 1 << 0 // Subvolume read-only.
)
