# Btrfs-Read Architecture

Technical architecture documentation for the Btrfs read-only filesystem implementation in Go.

## Table of Contents

1. [Overview](#overview)
2. [Architecture Design](#architecture-design)
3. [Core Modules](#core-modules)
4. [Data Flow](#data-flow)
5. [Key Data Structures](#key-data-structures)
6. [Implementation Details](#implementation-details)

---

## Overview

### Project Goals

Develop a read-only Btrfs filesystem implementation that can directly read and parse Btrfs structures from block devices, providing file and directory access capabilities.

### Core Features

- ✅ Read-only Btrfs filesystem access
- ✅ Basic file reading and directory traversal
- ✅ Logical to physical address mapping (Chunk Tree)
- ✅ B-Tree index traversal
- ✅ Support for INLINE and REGULAR file types
- ✅ Checksum verification (CRC32C)
- ✅ Multi-level directory support
- ❌ No write operations
- ❌ No compression support (future)
- ❌ No encryption support

### Reference

This project's architecture is inspired by [btrfs-fuse](https://github.com/adam900710/btrfs-fuse), a mature userspace read-only Btrfs implementation.

---

## Architecture Design

### Five-Layer Architecture

See diagram: [diagrams/architecture.md](../diagrams/architecture.md)

```
┌─────────────────────────────────┐
│   Application Layer (CLI)       │  ← Command-line interface
├─────────────────────────────────┤
│   Filesystem Layer (pkg/fs)     │  ← File/directory operations
├─────────────────────────────────┤
│   B-Tree Layer (pkg/btree)      │  ← Index search
├─────────────────────────────────┤
│   Chunk Layer (pkg/chunk)       │  ← Address mapping
├─────────────────────────────────┤
│   Device Layer (pkg/device)     │  ← Physical I/O
└─────────────────────────────────┘
```

### Layer Responsibilities

| Layer | Responsibilities |
|-------|------------------|
| **Application** | User interface (CLI) |
| **Filesystem** | File and directory operations, path resolution |
| **B-Tree** | Metadata indexing and search, key comparison |
| **Chunk** | Logical→Physical address mapping, RAID handling |
| **Device** | Physical I/O, block caching, checksum verification |

---

## Core Modules

### 1. Device Layer (`pkg/device`)

**Purpose:** Physical device I/O and caching

**Key Components:**
- `device.go` - Block device operations
- `cache.go` - LRU block cache (256 blocks)
- `super.go` - Superblock reading

**Features:**
- ReadAt interface for direct I/O
- LRU cache to reduce disk reads
- Multiple superblock support (primary + backups)

### 2. Chunk Layer (`pkg/chunk`)

**Purpose:** Logical to physical address translation

**Key Components:**
- `manager.go` - Chunk mapping management
- `loader.go` - Chunk tree loading
- `chunk.go` - Chunk structure definitions

**Features:**
- Red-black tree for fast chunk lookup
- RAID type handling (SINGLE, DUP)
- Stripe calculations

**Address Mapping Flow:**
See diagram: [diagrams/address-mapping.md](../diagrams/address-mapping.md)

1. Receive logical address
2. Search chunk in red-black tree
3. Calculate offset within chunk
4. Apply RAID type logic
5. Return physical address

### 3. B-Tree Layer (`pkg/btree`)

**Purpose:** Metadata indexing and search

**Key Components:**
- `search.go` - B-Tree search algorithm
- `node.go` - Node structure parsing

**Features:**
- Recursive tree traversal
- Binary search within nodes
- Path tracking (root → leaf)

**Search Flow:**
See diagram: [diagrams/btree-search.md](../diagrams/btree-search.md)

1. Start from root node
2. Parse node header and items
3. Binary search for target key
4. If internal node: recurse to child
5. If leaf node: return result

### 4. Filesystem Layer (`pkg/fs`)

**Purpose:** High-level file and directory operations

**Key Components:**
- `filesystem.go` - Filesystem implementation

**Features:**
- Path resolution (multi-level support)
- Directory listing
- File reading (INLINE and REGULAR types)
- DIR_ITEM and INODE_ITEM lookup

**File Read Flow:**
See diagram: [diagrams/file-read-flow.md](../diagrams/file-read-flow.md)

1. Path resolution → find inode
2. Read INODE_ITEM → get metadata
3. Find EXTENT_DATA → locate file data
4. Map logical → physical address
5. Read and return data

### 5. Application Layer (`cmd/btrfs-read`)

**Purpose:** Command-line interface

**Commands:**
- `info` - Show superblock information
- `ls` - List directory contents
- `cat` - Read file content

**Features:**
- JSON output support
- Configurable log levels
- Error handling and user-friendly messages

---

## Data Flow

### Initialization Flow

See diagram: [diagrams/init-flow.md](../diagrams/init-flow.md)

1. **Device Scan** - Open block device(s)
2. **Read Superblock** - Load primary (+ backups if needed)
3. **Load Chunk Tree** - Build logical→physical mapping
4. **Locate FS Tree** - Find filesystem tree root
5. **Initialize Cache** - Set up LRU cache

### File Read Flow

1. **Path Resolution**
   - Split path into components
   - Start from root directory (inode 256)
   - For each component:
     - Calculate CRC32C hash
     - Search DIR_ITEM in FS Tree
     - Get next inode number

2. **Inode Lookup**
   - Search INODE_ITEM with inode number
   - Get file metadata (size, mode, etc.)

3. **Extent Lookup**
   - Search EXTENT_DATA items
   - Handle INLINE or REGULAR types

4. **Data Reading**
   - Map logical to physical address
   - Read from device
   - Verify checksum (if applicable)
   - Return data

---

## Key Data Structures

### Superblock

Primary metadata located at offset 0x10000 (64KB):

```go
type Superblock struct {
    Magic         [8]byte   // "_BHRfS_M"
    FSID          [16]byte  // Filesystem UUID
    ByteNr        uint64    // Physical address
    Flags         uint64
    Generation    uint64
    Root          uint64    // Root tree logical address
    ChunkRoot     uint64    // Chunk tree logical address
    LogRoot       uint64    // Log tree logical address
    TotalBytes    uint64    // Filesystem size
    BytesUsed     uint64    // Used space
    NodeSize      uint32    // Node size (typically 16KB)
    SectorSize    uint32    // Sector size (typically 4KB)
    // ... more fields
}
```

### B-Tree Node Header

```go
type BtrfsHeader struct {
    Checksum    [32]byte  // CRC32C checksum
    FSID        [16]byte  // Filesystem UUID
    ByteNr      uint64    // Logical address
    Flags       uint64
    ChunkTreeUUID [16]byte
    Generation  uint64
    Owner       uint64    // Tree ID
    NrItems     uint32    // Number of items
    Level       uint8     // Node level (0 = leaf)
}
```

### Btrfs Key

Used for B-Tree indexing:

```go
type BtrfsKey struct {
    ObjectID uint64  // Object identifier (e.g., inode number)
    Type     uint8   // Key type (INODE_ITEM, DIR_ITEM, etc.)
    Offset   uint64  // Type-specific offset
}
```

**Key Comparison:**
1. Compare ObjectID
2. If equal, compare Type
3. If equal, compare Offset

### Chunk Mapping

```go
type ChunkMapping struct {
    LogicalStart uint64    // Logical address start
    Length       uint64    // Chunk length
    PhysicalStart uint64   // Physical address start
    DeviceID     uint64    // Device ID
    Type         uint64    // RAID type flags
    Stripes      []Stripe  // Stripe information
}
```

---

## Implementation Details

### Multi-Level Path Traversal

```go
func (fs *FileSystem) lookupPath(path string) (uint64, error) {
    // Start from root directory
    currentIno := uint64(256)
    
    // Split path and traverse
    parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
    
    for _, part := range parts {
        if part == "" {
            continue
        }
        
        // Lookup DIR_ITEM with CRC32C hash
        hash := crc32c(part)
        ino, err := fs.lookupDirItem(currentIno, part, hash)
        if err != nil {
            return 0, err
        }
        
        currentIno = ino
    }
    
    return currentIno, nil
}
```

### B-Tree Binary Search

```go
func binarySearchNode(node *Node, key *BtrfsKey) int {
    left, right := 0, int(node.Header.NrItems)
    
    for left < right {
        mid := (left + right) / 2
        cmp := node.Items[mid].Key.Compare(key)
        
        if cmp < 0 {
            left = mid + 1
        } else {
            right = mid
        }
    }
    
    return left
}
```

### LRU Cache

- **Size:** 256 blocks (configurable)
- **Eviction:** Least Recently Used
- **Thread-safe:** Mutex-protected
- **Hit ratio:** Typically 70-80% for sequential reads

### Logging System

Four levels: Debug, Info, Warn, Error

```go
// Set log level
logger.SetLevelFromString("debug")

// Log messages
logger.Debug("FS Tree root: 0x%x", fsTreeRoot)
logger.Info("File read successfully")
logger.Warn("Chunk parsing issue")
logger.Error("Data corruption detected")
```

Logs go to stderr, output goes to stdout for easy separation.

---

## Performance Characteristics

### Time Complexity

| Operation | Complexity | Notes |
|-----------|-----------|-------|
| Superblock read | O(1) | Fixed offset |
| Chunk lookup | O(log n) | Red-black tree |
| B-Tree search | O(log n) | Tree height typically 2-4 |
| Path resolution | O(d × log n) | d = path depth |
| File read | O(1) + O(e) | e = number of extents |

### Space Complexity

| Component | Size | Notes |
|-----------|------|-------|
| LRU cache | ~4MB | 256 × 16KB blocks |
| Chunk tree | ~1KB per chunk | Red-black tree nodes |
| B-Tree path | ~100KB | Max depth × node size |

### Optimization Strategies

1. **LRU Caching** - Reduce duplicate disk reads
2. **Lazy Loading** - Load data on demand
3. **Binary Search** - Fast key lookup in nodes
4. **Direct I/O** - No filesystem overhead

---

## Testing Strategy

### Unit Tests

- Superblock parsing
- Key comparison
- Hash calculation
- Data structure marshaling

### Integration Tests

- Read from real Btrfs images
- Multi-level directory traversal
- Various file types (INLINE, REGULAR)
- Error handling

### Test Images

Created with scripts in `tests/`:
- `test.img` - Basic filesystem
- `test_multilevel.img` - 4-level deep directories

---

## Reference

### Btrfs Specifications

- [Btrfs Wiki](https://btrfs.wiki.kernel.org/)
- [On-disk Format](https://btrfs.wiki.kernel.org/index.php/On-disk_Format)
- [Btrfs Design](https://btrfs.wiki.kernel.org/index.php/Btrfs_design)

### Related Projects

- [btrfs-fuse](https://github.com/adam900710/btrfs-fuse) - Reference implementation
- [btrfs-progs](https://github.com/kdave/btrfs-progs) - Official Btrfs tools

---

## Module Dependency Graph

```
cmd/btrfs-read
    ↓
pkg/fs (filesystem operations)
    ↓
pkg/btree (B-Tree search)
    ↓
pkg/chunk (address mapping)
    ↓
pkg/device (physical I/O)
    ↓
pkg/ondisk (structure definitions)
```

All modules depend on:
- `pkg/logger` - Logging
- `pkg/errors` - Error handling
