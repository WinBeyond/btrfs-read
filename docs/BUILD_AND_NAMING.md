# Build and Naming Configuration

## Executable Name
- **Binary Name**: `btrfs-read`
- **Previous Name**: `btrfs-cli` (deprecated)

## Build Directory Structure
```
btrfs_read/
├── build/                  # Build output directory
│   └── btrfs-read         # Main executable (3.0M)
├── cmd/
│   └── btrfs-read/        # CLI source code
│       └── main.go
└── ...
```

## Building the Project

### Using Make (Recommended)
```bash
# Clean and build
make clean
make build

# The executable will be in: build/btrfs-read
```

### Manual Build
```bash
# Create build directory
mkdir -p build

# Build executable
go build -o build/btrfs-read ./cmd/btrfs-read
```

## Running the Executable

### From Project Root
```bash
# Show help
./build/btrfs-read

# List directory
./build/btrfs-read ls tests/testdata/test.img /

# Read file
./build/btrfs-read cat tests/testdata/test.img /hello.txt

# JSON output
./build/btrfs-read ls --json tests/testdata/test.img /
./build/btrfs-read cat --json tests/testdata/test.img /hello.txt
```

## Available Commands

### info
Show superblock information
```bash
./build/btrfs-read info <image>
```

### ls
List directory contents
```bash
./build/btrfs-read ls [--json] <image> [path]

# Examples:
./build/btrfs-read ls tests/testdata/test.img /
./build/btrfs-read ls tests/testdata/test.img /level1/level2
./build/btrfs-read ls --json tests/testdata/test.img /
```

### cat
Read file content
```bash
./build/btrfs-read cat [--json] <image> <path>

# Examples:
./build/btrfs-read cat tests/testdata/test.img /hello.txt
./build/btrfs-read cat tests/testdata/test.img /a/b/c/deep.txt
./build/btrfs-read cat --json tests/testdata/test.img /readme.txt
```

## Makefile Targets

```bash
make build              # Build CLI tool
make test               # Run all tests
make test-unit          # Run unit tests only
make test-integration   # Run integration tests
make clean              # Clean build artifacts
make fmt                # Format code
make vet                # Run go vet
make help               # Show all targets
```

## Installation

### System-wide Installation
```bash
make install
# Installs to $GOPATH/bin/btrfs-read
```

### Manual Installation
```bash
sudo cp build/btrfs-read /usr/local/bin/
```

## Multi-Level Directory Support

The tool supports arbitrarily deep directory structures:

```bash
# Navigate through multiple levels
./build/btrfs-read ls tests/testdata/test_multilevel.img /level1/level2/level3/level4

# Read files at any depth
./build/btrfs-read cat tests/testdata/test_multilevel.img /a/b/c/deep.txt
```

## JSON Output Format

### Directory Listing
```json
{
  "path": "/level1/level2",
  "entries": [
    {
      "name": "file.txt",
      "inode": 257,
      "type": 1,
      "is_dir": false
    }
  ]
}
```

### File Reading
```json
{
  "path": "/hello.txt",
  "size": 18,
  "content": "Hello from Btrfs!\n"
}
```

## Testing

### Run Demo Scripts
```bash
# Basic demo
./demo.sh

# Multi-level directory test
./test_multilevel.sh

# Setup verification
./verify_setup.sh
```

## Clean Build Environment

To ensure a clean build environment:
```bash
# Remove all build artifacts
make clean

# Rebuild from scratch
make build

# Verify setup
./verify_setup.sh
```

## Notes

- The executable is now always built in the `build/` directory
- Root directory should not contain any executables
- All documentation references `./build/btrfs-read`
- Makefile automatically creates the build directory
- The old name `btrfs-cli` has been fully deprecated
