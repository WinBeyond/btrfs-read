# Btrfs-Read

A read-only Btrfs filesystem implementation in Go. Read files and directories from Btrfs images or devices without mounting.

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## Features

- ✅ Read Btrfs superblock information
- ✅ List directory contents (multi-level support)
- ✅ Read file contents at any depth
- ✅ JSON output format
- ✅ Support for INLINE and REGULAR file types
- ✅ Complete B-Tree traversal
- ✅ Chunk logical-to-physical address mapping
- ✅ Configurable logging levels

## Installation

### Using go install (Recommended)

```bash
go install github.com/WinBeyond/btrfs-read/cmd/btrfs-read@latest
```

Make sure `$GOPATH/bin` is in your `PATH`:
```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

### From Source

```bash
git clone https://github.com/WinBeyond/btrfs-read.git
cd btrfs-read
make build
```

## Quick Start

```bash
# Show filesystem information
btrfs-read info <image>

# List directory contents
btrfs-read ls <image> [path]

# Read file content
btrfs-read cat <image> <path>

# JSON output
btrfs-read ls --json <image> /
btrfs-read cat --json <image> /file.txt

# Set log level
btrfs-read ls -l debug <image> /
```

## Usage Examples

```bash
# Create a test image (requires root)
sudo bash tests/create-test-image.sh

# List root directory
btrfs-read ls tests/testdata/test.img /

# Read a file
btrfs-read cat tests/testdata/test.img /hello.txt

# Navigate multi-level directories
btrfs-read ls tests/testdata/test.img /dir1/dir2/dir3
btrfs-read cat tests/testdata/test.img /a/b/c/file.txt

# JSON output
btrfs-read ls --json tests/testdata/test.img /
```

## Commands

### info
Show Btrfs superblock information

```bash
btrfs-read info <image>
```

### ls
List directory contents

```bash
btrfs-read ls [--json] [-l level] <image> [path]

# Options:
#   --json        Output in JSON format
#   -l, --log-level   Set log level (debug, info, warn, error)
```

### cat
Read file content

```bash
btrfs-read cat [--json] [-l level] <image> <path>

# Options:
#   --json        Output in JSON format
#   -l, --log-level   Set log level (debug, info, warn, error)
```

## Log Levels

Control verbosity with the `-l` or `--log-level` flag:

- `debug` - Detailed debugging information
- `info` - Normal operation (default)
- `warn` - Warnings only
- `error` - Errors only

```bash
# Debug mode
btrfs-read ls -l debug tests/testdata/test.img /

# Quiet mode
btrfs-read cat -l error tests/testdata/test.img /file.txt
```

## Project Structure

```
btrfs_read/
├── cmd/btrfs-read/     # CLI application
├── pkg/                # Core packages
│   ├── btree/         # B-Tree implementation
│   ├── chunk/         # Chunk management
│   ├── device/        # Device operations
│   ├── fs/            # Filesystem layer
│   ├── logger/        # Logging system
│   ├── ondisk/        # On-disk format definitions
│   └── errors/        # Error definitions
├── tests/             # Test files and images
├── docs/              # Documentation
├── scripts/           # Utility scripts
└── diagrams/          # Architecture diagrams
```

## Architecture

Five-layer design:

1. **Application Layer** - CLI interface
2. **Filesystem Layer** - File and directory operations
3. **B-Tree Layer** - Metadata indexing and search
4. **Chunk Layer** - Logical to physical address mapping
5. **Device Layer** - Physical I/O and caching

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for details.

## Development

### Build

```bash
make build          # Build binary
make clean          # Clean build artifacts
make fmt            # Format code
make vet            # Run go vet
```

### Test

```bash
make test           # Run all tests
make test-unit      # Unit tests only
make test-integration  # Integration tests only
```

### Scripts

```bash
./scripts/demo.sh              # Basic demo
./scripts/test_multilevel.sh   # Multi-level directory test
./scripts/final_test.sh        # Comprehensive test
./scripts/verify_setup.sh      # Verify setup
```

## Limitations

- ❌ Read-only (no write support)
- ❌ No compression support
- ❌ No encryption support
- ❌ No snapshot/subvolume switching
- ✅ INLINE and REGULAR file types supported
- ✅ Single-device Btrfs only
- ✅ Multi-level directory traversal

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

## Documentation

- [Architecture](docs/ARCHITECTURE.md) - Technical architecture and design
- [Usage](docs/USAGE.md) - Detailed usage guide
- [Diagrams](diagrams/) - Architecture diagrams (Mermaid format)

## License

MIT License

## References

- [Btrfs Wiki](https://btrfs.wiki.kernel.org/)
- [Btrfs On-disk Format](https://btrfs.wiki.kernel.org/index.php/On-disk_Format)
- [btrfs-fuse](https://github.com/adam900710/btrfs-fuse) - Reference implementation
