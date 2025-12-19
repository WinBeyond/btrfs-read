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

// FileSystem represents a Btrfs filesystem.
type FileSystem struct {
	superblock   *ondisk.Superblock
	device       *device.FileDevice
	chunkManager *chunk.Manager
	cache        *device.BlockCache

	fsTreeRoot    uint64
	btreeSearcher *btree.Searcher
}

// Open opens a filesystem.
func Open(devicePath string) (*FileSystem, error) {
	// 1. Open device.
	dev, err := device.NewFileDevice(devicePath)
	if err != nil {
		return nil, errors.Wrap("FileSystem.Open", err)
	}

	// 2. Read superblock.
	sbReader := device.NewSuperblockReader(dev)
	sb, err := sbReader.ReadLatest()
	if err != nil {
		dev.Close()
		return nil, errors.Wrap("FileSystem.Open.ReadSuperblock", err)
	}

	// 3. Initialize chunk manager.
	chunkMgr := chunk.NewManager()

	// Initialize from the system chunk array (bootstrap chunks).
	if err := chunkMgr.ParseSystemChunkArray(sb.SysChunkArray[:], sb.SysChunkArraySize); err != nil {
		dev.Close()
		return nil, errors.Wrap("FileSystem.Open.ParseChunks", err)
	}

	// 4. Create cache.
	cache := device.NewBlockCache(256)

	// 5. Create filesystem instance (temporary, for loading the chunk tree).
	fs := &FileSystem{
		superblock:   sb,
		device:       dev,
		chunkManager: chunkMgr,
		cache:        cache,
		fsTreeRoot:   sb.Root,
	}

	// 6. Create B-Tree searcher.
	fs.btreeSearcher = btree.NewSearcher(fs, sb.NodeSize)

	// 7. Load all chunks from the chunk tree.
	loader := chunk.NewChunkTreeLoader(chunkMgr, fs, sb.NodeSize)
	if err := loader.LoadFromChunkTree(sb.ChunkRoot); err != nil {
		dev.Close()
		return nil, errors.Wrap("FileSystem.Open.LoadChunkTree", err)
	}

	// 8. Find the FS_TREE root from the Root Tree.
	fsTreeRoot, err := fs.findFSTreeRoot()
	if err != nil {
		dev.Close()
		return nil, errors.Wrap("FileSystem.Open.FindFSTree", err)
	}
	fs.fsTreeRoot = fsTreeRoot

	logger.Debug("FS Tree root: 0x%x", fsTreeRoot)

	return fs, nil
}

// findFSTreeRoot finds the FS_TREE root node address from the Root Tree.
func (fs *FileSystem) findFSTreeRoot() (uint64, error) {
	// Search ROOT_ITEM: objectid=5 (FS_TREE), type=132 (ROOT_ITEM_KEY).
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

	// Check whether the FS_TREE ROOT_ITEM was found.
	if item.Key.ObjectID != 5 || item.Key.Type != 132 {
		return 0, fmt.Errorf("FS_TREE ROOT_ITEM not found, got key: objectid=%d type=%d",
			item.Key.ObjectID, item.Key.Type)
	}

	// Parse ROOT_ITEM and extract bytenr (root node address).
	// ROOT_ITEM format (simplified):
	// ...
	// offset 176: bytenr (8 bytes)
	if len(item.Data) < 184 {
		return 0, fmt.Errorf("ROOT_ITEM data too short: %d bytes", len(item.Data))
	}

	bytenr := binary.LittleEndian.Uint64(item.Data[176:184])

	return bytenr, nil
}

// Close closes the filesystem.
func (fs *FileSystem) Close() error {
	if fs.device != nil {
		return fs.device.Close()
	}
	return nil
}

// ReadNode implements btree.NodeReader.
func (fs *FileSystem) ReadNode(logical uint64, nodeSize uint32) (*btree.Node, error) {
	// 1. Logical address -> physical address.
	physAddr, err := fs.chunkManager.LogicalToPhysical(logical)
	if err != nil {
		return nil, errors.Wrap("ReadNode.LogicalToPhysical", err)
	}

	// 2. Check cache.
	cacheKey := logical
	if cached, ok := fs.cache.Get(cacheKey); ok {
		return btree.UnmarshalNode(cached, nodeSize)
	}

	// 3. Read from device.
	buf := make([]byte, nodeSize)
	n, err := fs.device.ReadAt(buf, int64(physAddr.Offset))
	if err != nil || n != int(nodeSize) {
		return nil, fmt.Errorf("failed to read node: %w", err)
	}

	// 4. Put into cache.
	fs.cache.Put(cacheKey, buf)

	// 5. Parse node.
	return btree.UnmarshalNode(buf, nodeSize)
}

// DirEntry represents a directory entry.
type DirEntry struct {
	Name  string `json:"name"`
	Inode uint64 `json:"inode"`
	Type  uint8  `json:"type"`
	IsDir bool   `json:"is_dir"`
}

// ListDirectory lists directory contents.
func (fs *FileSystem) ListDirectory(path string) ([]*DirEntry, error) {
	// 1. Resolve path and get directory inode.
	var dirIno uint64
	if path == "/" || path == "" {
		dirIno = 256 // Root directory.
	} else {
		ino, err := fs.lookupPath(path)
		if err != nil {
			return nil, err
		}
		dirIno = ino
	}

	// 2. Iterate directory entries.
	// Search all DIR_INDEX or DIR_ITEM.
	// Simplified: only search DIR_INDEX (type=96).
	entries := make([]*DirEntry, 0)
	seenOffsets := make(map[uint64]bool)

	// Iterate using DIR_INDEX (faster).
	// DIR_INDEX key: objectid=dir_ino, type=96, offset=index.
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

		// Check whether it is still the same directory's DIR_INDEX.
		if item.Key.ObjectID != dirIno || item.Key.Type != 96 {
			break
		}

		// Deduplicate: skip offsets already seen.
		if seenOffsets[item.Key.Offset] {
			continue
		}
		seenOffsets[item.Key.Offset] = true

		// If the found offset is greater than current index, jump to it.
		if item.Key.Offset > index {
			index = item.Key.Offset
		}

		// Parse DIR_INDEX.
		entry, err := fs.parseDirIndex(item.Data)
		if err != nil {
			continue
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// parseDirIndex parses DIR_INDEX data.
func (fs *FileSystem) parseDirIndex(data []byte) (*DirEntry, error) {
	if len(data) < 30 {
		return nil, fmt.Errorf("DIR_INDEX data too short")
	}

	// DIR_INDEX/DIR_ITEM format:
	// location (key: 17 bytes) + transid (8) + data_len (2) + name_len (2) + type (1) + name

	// location.objectid (target inode).
	targetIno := binary.LittleEndian.Uint64(data[0:8])

	// name_len.
	nameLen := binary.LittleEndian.Uint16(data[27:29])

	// type.
	fileType := data[29]

	// name.
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

// ReadFile reads file contents.
func (fs *FileSystem) ReadFile(path string) ([]byte, error) {
	// 1. Resolve path and get inode.
	ino, err := fs.lookupPath(path)
	if err != nil {
		return nil, err
	}

	// 2. Read inode info.
	inodeInfo, err := fs.readInode(ino)
	if err != nil {
		return nil, err
	}

	if inodeInfo.Size == 0 {
		return []byte{}, nil
	}

	// 3. Read file data.
	return fs.readFileData(ino, inodeInfo.Size)
}

// lookupPath resolves a path (simplified: assumes /filename).
func (fs *FileSystem) lookupPath(path string) (uint64, error) {
	// Support multi-level traversal.
	// Path format: /dir1/dir2/filename.

	if !strings.HasPrefix(path, "/") {
		return 0, fmt.Errorf("path must start with /")
	}

	path = strings.TrimPrefix(path, "/")

	if path == "" {
		// Root directory.
		return 256, nil // BTRFS_FIRST_FREE_OBJECTID
	}

	// Split the path.
	parts := strings.Split(path, "/")

	// Start from the root directory.
	currentIno := uint64(256)

	for i, part := range parts {
		if part == "" {
			continue
		}

		// Look for the next component in the current directory.
		ino, err := fs.lookupDirItem(currentIno, part)
		if err != nil {
			return 0, fmt.Errorf("file not found: %s", path)
		}

		// If not the last component, check that it is a directory.
		if i < len(parts)-1 {
			// Should be a directory, continue to the next level.
			currentIno = ino
		} else {
			// Last component; return its inode.
			return ino, nil
		}
	}

	return currentIno, nil
}

// lookupDirItem finds an entry in a directory.
func (fs *FileSystem) lookupDirItem(dirIno uint64, name string) (uint64, error) {
	// Compute the name hash.
	nameHash := crc32Hash([]byte(name))

	// Search DIR_ITEM.
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

	// Check for an exact match.
	if item.Key.Compare(key) != 0 {
		return 0, fmt.Errorf("file not found: %s", name)
	}

	// Parse DIR_ITEM and extract the target inode.
	if len(item.Data) < 30 {
		return 0, fmt.Errorf("DIR_ITEM data too short")
	}

	// DIR_ITEM format: location(key:17) + transid(8) + data_len(2) + name_len(2) + type(1) + name.
	targetIno := binary.LittleEndian.Uint64(item.Data[0:8])

	return targetIno, nil
}

// InodeInfo holds inode information.
type InodeInfo struct {
	Ino  uint64
	Size uint64
	Mode uint32
}

// readInode reads an inode.
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

	// Parse INODE_ITEM.
	if len(item.Data) < 160 {
		return nil, fmt.Errorf("INODE_ITEM data too short")
	}

	// INODE_ITEM format: generation(8) + transid(8) + size(8) + ...
	size := binary.LittleEndian.Uint64(item.Data[16:24])
	mode := binary.LittleEndian.Uint32(item.Data[52:56])

	return &InodeInfo{
		Ino:  ino,
		Size: size,
		Mode: mode,
	}, nil
}

// readFileData reads file data.
func (fs *FileSystem) readFileData(ino uint64, size uint64) ([]byte, error) {
	// Find EXTENT_DATA.
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

	// Might not be an exact match, but should be the first EXTENT_DATA.
	if item.Key.ObjectID != ino || item.Key.Type != 108 {
		return nil, fmt.Errorf("extent data not found for inode %d", ino)
	}

	// Parse EXTENT_DATA.
	if len(item.Data) < 21 {
		return nil, fmt.Errorf("EXTENT_DATA too short")
	}

	// EXTENT_DATA format: generation(8) + ram_bytes(8) + compression(1) + encryption(1) + other(2) + type(1).
	extentType := item.Data[20]

	if extentType == 0 { // INLINE
		// Data is embedded in the item.
		dataStart := 21
		if dataStart >= len(item.Data) {
			return []byte{}, nil
		}
		return item.Data[dataStart:], nil
	} else if extentType == 1 { // REGULAR
		// Need to read a separate extent.
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

		// Read actual data.
		return fs.readExtent(diskBytenr, diskNumBytes, numBytes)
	}

	return nil, fmt.Errorf("unsupported extent type: %d", extentType)
}

// readExtent reads extent data.
func (fs *FileSystem) readExtent(logical uint64, diskSize uint64, dataSize uint64) ([]byte, error) {
	// Logical address -> physical address.
	physAddr, err := fs.chunkManager.LogicalToPhysical(logical)
	if err != nil {
		return nil, err
	}

	// Read data.
	buf := make([]byte, diskSize)
	n, err := fs.device.ReadAt(buf, int64(physAddr.Offset))
	if err != nil || uint64(n) != diskSize {
		return nil, fmt.Errorf("failed to read extent: %w", err)
	}

	// Simplified: assume no compression.
	if dataSize < diskSize {
		return buf[:dataSize], nil
	}

	return buf, nil
}

// crc32Hash computes CRC32 hash (for DIR_ITEM).
func crc32Hash(data []byte) uint64 {
	// Btrfs uses crc32c with seed ~1
	// Implement it manually to match kernel's crc32c exactly
	crc := ^uint32(1) // ~1 = 0xFFFFFFFE

	for _, b := range data {
		crc = crc32cTable[(crc^uint32(b))&0xFF] ^ (crc >> 8)
	}

	return uint64(crc)
}

// CRC32 table (simplified).
