# Architecture Diagrams

This directory contains architecture diagrams and flowcharts for the Btrfs-Read project, written in Mermaid format.

## Diagrams

### 1. System Architecture
**File**: [architecture.md](architecture.md)

Five-layer architecture design:
- Application Layer
- Filesystem Layer
- B-Tree Layer
- Chunk Layer (Logical Block Layer)
- Device Layer (Physical Block Layer)

### 2. Address Mapping Flow
**File**: [address-mapping.md](address-mapping.md)

Logical to physical address mapping process:
- Chunk lookup in red-black tree
- RAID type handling (SINGLE/DUP/RAID0/1/5/6/10)
- Physical address calculation

### 3. B-Tree Search Flow
**File**: [btree-search.md](btree-search.md)

B-Tree index search algorithm:
- Node reading and parsing
- Binary search implementation
- Internal node vs leaf node handling
- Path tracking

### 4. File Read Flow
**File**: [file-read-flow.md](file-read-flow.md)

Complete file reading sequence diagram:
1. Path resolution (directory tree traversal)
2. Inode information lookup
3. Extent lookup
4. Data reading and decompression
5. Checksum verification

### 5. Initialization Flow
**File**: [init-flow.md](init-flow.md)

Filesystem initialization process:
1. Device scanning
2. Superblock reading
3. Chunk Tree initialization
4. FS Tree location
5. Cache initialization

## Viewing Diagrams

### On GitHub
GitHub natively supports Mermaid syntax. Just click on any `.md` file to view the rendered diagram.

### Locally

#### Method 1: VS Code
1. Install "Markdown Preview Mermaid Support" extension
2. Open `.md` file
3. Press `Ctrl+Shift+V` (Windows/Linux) or `Cmd+Shift+V` (Mac) to preview

#### Method 2: Mermaid Live Editor
1. Visit https://mermaid.live/
2. Copy diagram code
3. Paste into editor to view

#### Method 3: Mermaid CLI
```bash
# Install mermaid-cli
npm install -g @mermaid-js/mermaid-cli

# Generate PNG image
mmdc -i architecture.md -o architecture.png

# Generate SVG image
mmdc -i architecture.md -o architecture.svg
```

## Diagram Format

All diagrams use Mermaid format with these advantages:
- ✅ Plain text, easy version control
- ✅ GitHub native support
- ✅ Convertible to multiple formats (PNG, SVG, PDF)
- ✅ Easy to maintain and update

## Related Documentation

- [ARCHITECTURE.md](../docs/ARCHITECTURE.md) - Detailed architecture documentation
- [README.md](../README.md) - Project overview
- [USAGE.md](../docs/USAGE.md) - Usage guide

## Contributing

To update diagrams:
1. Edit the corresponding `.md` file
2. Modify the Mermaid code block
3. Preview locally to verify
4. Submit a pull request
