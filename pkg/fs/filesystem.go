package fs

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"strings"

	"github.com/WinBeyond/btrfs-read/pkg/btree"
	"github.com/WinBeyond/btrfs-read/pkg/chunk"
	"github.com/WinBeyond/btrfs-read/pkg/device"
	"github.com/WinBeyond/btrfs-read/pkg/errors"
	"github.com/WinBeyond/btrfs-read/pkg/logger"
	"github.com/WinBeyond/btrfs-read/pkg/ondisk"
)

// CRC32C table (Castagnoli polynomial) for btrfs name hashing
var crc32cTable = crc32.MakeTable(crc32.Castagnoli)

// FileSystem Btrfs 文件系统
type FileSystem struct {
	superblock   *ondisk.Superblock
	device       *device.FileDevice
	chunkManager *chunk.Manager
	cache        *device.BlockCache

	fsTreeRoot    uint64
	btreeSearcher *btree.Searcher
}

// Open 打开文件系统
func Open(devicePath string) (*FileSystem, error) {
	// 1. 打开设备
	dev, err := device.NewFileDevice(devicePath)
	if err != nil {
		return nil, errors.Wrap("FileSystem.Open", err)
	}

	// 2. 读取 superblock
	sbReader := device.NewSuperblockReader(dev)
	sb, err := sbReader.ReadLatest()
	if err != nil {
		dev.Close()
		return nil, errors.Wrap("FileSystem.Open.ReadSuperblock", err)
	}

	// 3. 初始化 chunk manager
	chunkMgr := chunk.NewManager()

	// 从系统 chunk 数组初始化（bootstrap chunks）
	if err := chunkMgr.ParseSystemChunkArray(sb.SysChunkArray[:], sb.SysChunkArraySize); err != nil {
		dev.Close()
		return nil, errors.Wrap("FileSystem.Open.ParseChunks", err)
	}

	// 4. 创建缓存
	cache := device.NewBlockCache(256)

	// 5. 创建文件系统实例（临时的，用于加载 chunk tree）
	fs := &FileSystem{
		superblock:   sb,
		device:       dev,
		chunkManager: chunkMgr,
		cache:        cache,
		fsTreeRoot:   sb.Root,
	}

	// 6. 创建 B-Tree searcher
	fs.btreeSearcher = btree.NewSearcher(fs, sb.NodeSize)

	// 7. 从 chunk tree 加载所有 chunks
	loader := chunk.NewChunkTreeLoader(chunkMgr, fs, sb.NodeSize)
	if err := loader.LoadFromChunkTree(sb.ChunkRoot); err != nil {
		dev.Close()
		return nil, errors.Wrap("FileSystem.Open.LoadChunkTree", err)
	}

	// 8. 从 Root Tree 查找 FS_TREE 的根
	fsTreeRoot, err := fs.findFSTreeRoot()
	if err != nil {
		dev.Close()
		return nil, errors.Wrap("FileSystem.Open.FindFSTree", err)
	}
	fs.fsTreeRoot = fsTreeRoot

	logger.Debug("FS Tree root: 0x%x", fsTreeRoot)

	return fs, nil
}

// findFSTreeRoot 从 Root Tree 查找 FS_TREE 的根节点地址
func (fs *FileSystem) findFSTreeRoot() (uint64, error) {
	// 搜索 ROOT_ITEM: objectid=5 (FS_TREE), type=132 (ROOT_ITEM_KEY)
	key := &btree.Key{
		ObjectID: 5,   // BTRFS_FS_TREE_OBJECTID
		Type:     132, // BTRFS_ROOT_ITEM_KEY
		Offset:   0,
	}

	path, err := fs.btreeSearcher.Search(fs.superblock.Root, key)
	if err != nil {
		return 0, fmt.Errorf("search failed: %w", err)
	}

	item, err := path.GetItem()
	if err != nil {
		return 0, fmt.Errorf("get item failed: %w", err)
	}

	// 检查是否找到了 FS_TREE 的 ROOT_ITEM
	if item.Key.ObjectID != 5 || item.Key.Type != 132 {
		return 0, fmt.Errorf("FS_TREE ROOT_ITEM not found, got key: objectid=%d type=%d",
			item.Key.ObjectID, item.Key.Type)
	}

	// 解析 ROOT_ITEM，获取 bytenr（根节点地址）
	// ROOT_ITEM 格式（简化）：
	// ...
	// offset 176: bytenr (8 bytes)
	if len(item.Data) < 184 {
		return 0, fmt.Errorf("ROOT_ITEM data too short: %d bytes", len(item.Data))
	}

	bytenr := binary.LittleEndian.Uint64(item.Data[176:184])

	return bytenr, nil
}

// Close 关闭文件系统
func (fs *FileSystem) Close() error {
	if fs.device != nil {
		return fs.device.Close()
	}
	return nil
}

// ReadNode 实现 btree.NodeReader 接口
func (fs *FileSystem) ReadNode(logical uint64, nodeSize uint32) (*btree.Node, error) {
	// 1. 逻辑地址 → 物理地址
	physAddr, err := fs.chunkManager.LogicalToPhysical(logical)
	if err != nil {
		return nil, errors.Wrap("ReadNode.LogicalToPhysical", err)
	}

	// 2. 检查缓存
	cacheKey := logical
	if cached, ok := fs.cache.Get(cacheKey); ok {
		return btree.UnmarshalNode(cached, nodeSize)
	}

	// 3. 从设备读取
	buf := make([]byte, nodeSize)
	n, err := fs.device.ReadAt(buf, int64(physAddr.Offset))
	if err != nil || n != int(nodeSize) {
		return nil, fmt.Errorf("failed to read node: %w", err)
	}

	// 4. 放入缓存
	fs.cache.Put(cacheKey, buf)

	// 5. 解析节点
	return btree.UnmarshalNode(buf, nodeSize)
}

// DirEntry 目录项
type DirEntry struct {
	Name  string `json:"name"`
	Inode uint64 `json:"inode"`
	Type  uint8  `json:"type"`
	IsDir bool   `json:"is_dir"`
}

// ListDirectory 列出目录内容
func (fs *FileSystem) ListDirectory(path string) ([]*DirEntry, error) {
	// 1. 路径解析，获取目录的 inode
	var dirIno uint64
	if path == "/" || path == "" {
		dirIno = 256 // 根目录
	} else {
		ino, err := fs.lookupPath(path)
		if err != nil {
			return nil, err
		}
		dirIno = ino
	}

	// 2. 遍历目录项
	// 搜索所有 DIR_INDEX 或 DIR_ITEM
	// 简化版：只搜索 DIR_INDEX (type=96)
	entries := make([]*DirEntry, 0)
	seenOffsets := make(map[uint64]bool)

	// 使用 DIR_INDEX 遍历（更快）
	// DIR_INDEX key: objectid=dir_ino, type=96, offset=index
	for index := uint64(0); index < 10000; index++ {
		key := &btree.Key{
			ObjectID: dirIno,
			Type:     96, // DIR_INDEX
			Offset:   index,
		}

		path, err := fs.btreeSearcher.Search(fs.fsTreeRoot, key)
		if err != nil {
			break
		}

		item, err := path.GetItem()
		if err != nil {
			break
		}

		// 检查是否还是同一个目录的 DIR_INDEX
		if item.Key.ObjectID != dirIno || item.Key.Type != 96 {
			break
		}

		// 去重：如果已经处理过这个offset，跳过
		if seenOffsets[item.Key.Offset] {
			continue
		}
		seenOffsets[item.Key.Offset] = true

		// 如果找到的offset大于当前index，跳到那个offset
		if item.Key.Offset > index {
			index = item.Key.Offset
		}

		// 解析 DIR_INDEX
		entry, err := fs.parseDirIndex(item.Data)
		if err != nil {
			continue
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// parseDirIndex 解析 DIR_INDEX 数据
func (fs *FileSystem) parseDirIndex(data []byte) (*DirEntry, error) {
	if len(data) < 30 {
		return nil, fmt.Errorf("DIR_INDEX data too short")
	}

	// DIR_INDEX/DIR_ITEM 格式：
	// location (key: 17 bytes) + transid (8) + data_len (2) + name_len (2) + type (1) + name

	// location.objectid (目标 inode)
	targetIno := binary.LittleEndian.Uint64(data[0:8])

	// name_len
	nameLen := binary.LittleEndian.Uint16(data[27:29])

	// type
	fileType := data[29]

	// name
	if len(data) < 30+int(nameLen) {
		return nil, fmt.Errorf("name data out of bounds")
	}
	name := string(data[30 : 30+nameLen])

	return &DirEntry{
		Name:  name,
		Inode: targetIno,
		Type:  fileType,
		IsDir: fileType == 2, // BTRFS_FT_DIR = 2
	}, nil
}

// ReadFile 读取文件内容
func (fs *FileSystem) ReadFile(path string) ([]byte, error) {
	// 1. 路径解析，获取 inode
	ino, err := fs.lookupPath(path)
	if err != nil {
		return nil, err
	}

	// 2. 读取 inode 信息
	inodeInfo, err := fs.readInode(ino)
	if err != nil {
		return nil, err
	}

	if inodeInfo.Size == 0 {
		return []byte{}, nil
	}

	// 3. 读取文件数据
	return fs.readFileData(ino, inodeInfo.Size)
}

// lookupPath 路径解析（简化版：假设路径是 /filename）
func (fs *FileSystem) lookupPath(path string) (uint64, error) {
	// 支持多层级路径遍历
	// 路径格式：/dir1/dir2/filename

	if !strings.HasPrefix(path, "/") {
		return 0, fmt.Errorf("path must start with /")
	}

	path = strings.TrimPrefix(path, "/")

	if path == "" {
		// 根目录
		return 256, nil // BTRFS_FIRST_FREE_OBJECTID
	}

	// 分割路径
	parts := strings.Split(path, "/")

	// 从根目录开始查找
	currentIno := uint64(256)

	for i, part := range parts {
		if part == "" {
			continue
		}

		// 在当前目录中查找下一级
		ino, err := fs.lookupDirItem(currentIno, part)
		if err != nil {
			return 0, fmt.Errorf("file not found: %s", path)
		}

		// 如果不是最后一个部分，检查它是否是目录
		if i < len(parts)-1 {
			// 应该是目录，继续下一级
			currentIno = ino
		} else {
			// 最后一个部分，返回其 inode
			return ino, nil
		}
	}

	return currentIno, nil
}

// lookupDirItem 在目录中查找项
func (fs *FileSystem) lookupDirItem(dirIno uint64, name string) (uint64, error) {
	// 计算名称的哈希
	nameHash := crc32Hash([]byte(name))

	// 搜索 DIR_ITEM
	key := &btree.Key{
		ObjectID: dirIno,
		Type:     84, // DIR_ITEM
		Offset:   nameHash,
	}

	path, err := fs.btreeSearcher.Search(fs.fsTreeRoot, key)
	if err != nil {
		logger.Debug("DIR_ITEM search failed: %v", err)
		return 0, err
	}

	item, err := path.GetItem()
	if err != nil {
		logger.Debug("Failed to get DIR_ITEM: %v", err)
		return 0, err
	}

	// 检查是否精确匹配
	if item.Key.Compare(key) != 0 {
		return 0, fmt.Errorf("file not found: %s", name)
	}

	// 解析 DIR_ITEM，提取目标 inode
	if len(item.Data) < 30 {
		return 0, fmt.Errorf("DIR_ITEM data too short")
	}

	// DIR_ITEM 格式：location(key:17) + transid(8) + data_len(2) + name_len(2) + type(1) + name
	targetIno := binary.LittleEndian.Uint64(item.Data[0:8])

	return targetIno, nil
}

// InodeInfo Inode 信息
type InodeInfo struct {
	Ino  uint64
	Size uint64
	Mode uint32
}

// readInode 读取 inode
func (fs *FileSystem) readInode(ino uint64) (*InodeInfo, error) {
	key := &btree.Key{
		ObjectID: ino,
		Type:     1, // INODE_ITEM
		Offset:   0,
	}

	path, err := fs.btreeSearcher.Search(fs.fsTreeRoot, key)
	if err != nil {
		return nil, err
	}

	item, err := path.GetItem()
	if err != nil {
		return nil, err
	}

	if item.Key.Compare(key) != 0 {
		return nil, fmt.Errorf("inode %d not found", ino)
	}

	// 解析 INODE_ITEM
	if len(item.Data) < 160 {
		return nil, fmt.Errorf("INODE_ITEM data too short")
	}

	// INODE_ITEM 格式：generation(8) + transid(8) + size(8) + ...
	size := binary.LittleEndian.Uint64(item.Data[16:24])
	mode := binary.LittleEndian.Uint32(item.Data[52:56])

	return &InodeInfo{
		Ino:  ino,
		Size: size,
		Mode: mode,
	}, nil
}

// readFileData 读取文件数据
func (fs *FileSystem) readFileData(ino uint64, size uint64) ([]byte, error) {
	// 查找 EXTENT_DATA
	key := &btree.Key{
		ObjectID: ino,
		Type:     108, // EXTENT_DATA
		Offset:   0,
	}

	path, err := fs.btreeSearcher.Search(fs.fsTreeRoot, key)
	if err != nil {
		return nil, err
	}

	item, err := path.GetItem()
	if err != nil {
		return nil, err
	}

	// 可能不是精确匹配，但应该是第一个 EXTENT_DATA
	if item.Key.ObjectID != ino || item.Key.Type != 108 {
		return nil, fmt.Errorf("extent data not found for inode %d", ino)
	}

	// 解析 EXTENT_DATA
	if len(item.Data) < 21 {
		return nil, fmt.Errorf("EXTENT_DATA too short")
	}

	// EXTENT_DATA 格式：generation(8) + ram_bytes(8) + compression(1) + encryption(1) + other(2) + type(1)
	extentType := item.Data[20]

	if extentType == 0 { // INLINE
		// 数据直接在 item 中
		dataStart := 21
		if dataStart >= len(item.Data) {
			return []byte{}, nil
		}
		return item.Data[dataStart:], nil
	} else if extentType == 1 { // REGULAR
		// 需要读取独立的 extent
		if len(item.Data) < 53 {
			return nil, fmt.Errorf("REGULAR extent data too short")
		}

		diskBytenr := binary.LittleEndian.Uint64(item.Data[21:29])
		diskNumBytes := binary.LittleEndian.Uint64(item.Data[29:37])
		// offset := binary.LittleEndian.Uint64(item.Data[37:45])
		numBytes := binary.LittleEndian.Uint64(item.Data[45:53])

		if diskBytenr == 0 {
			// Sparse file or hole
			return make([]byte, numBytes), nil
		}

		// 读取实际数据
		return fs.readExtent(diskBytenr, diskNumBytes, numBytes)
	}

	return nil, fmt.Errorf("unsupported extent type: %d", extentType)
}

// readExtent 读取 extent 数据
func (fs *FileSystem) readExtent(logical uint64, diskSize uint64, dataSize uint64) ([]byte, error) {
	// 逻辑地址 → 物理地址
	physAddr, err := fs.chunkManager.LogicalToPhysical(logical)
	if err != nil {
		return nil, err
	}

	// 读取数据
	buf := make([]byte, diskSize)
	n, err := fs.device.ReadAt(buf, int64(physAddr.Offset))
	if err != nil || uint64(n) != diskSize {
		return nil, fmt.Errorf("failed to read extent: %w", err)
	}

	// 简化：假设没有压缩
	if dataSize < diskSize {
		return buf[:dataSize], nil
	}

	return buf, nil
}

// crc32Hash 计算 CRC32 哈希（用于 DIR_ITEM）
func crc32Hash(data []byte) uint64 {
	// Btrfs uses crc32c with seed ~1
	// Implement it manually to match kernel's crc32c exactly
	crc := ^uint32(1) // ~1 = 0xFFFFFFFE

	for _, b := range data {
		crc = crc32cTable[(crc^uint32(b))&0xFF] ^ (crc >> 8)
	}

	return uint64(crc)
}

// CRC32 表（简化版）
