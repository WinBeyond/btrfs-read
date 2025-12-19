# Btrfs-Read Usage Guide

## Installation

### Method 1: Using go install

```bash
go install github.com/WinBeyond/btrfs-read/cmd/btrfs-read@latest

# Add to PATH if needed
export PATH=$PATH:$(go env GOPATH)/bin
```

### Method 2: From Source

```bash
git clone https://github.com/WinBeyond/btrfs-read.git
cd btrfs-read
make build
# Binary will be in: build/btrfs-read
```

## Commands

### info - Show Filesystem Information

Display Btrfs superblock information.

```bash
btrfs-read info <image>
```

**Example:**
```bash
btrfs-read info tests/testdata/test.img
```

**Output:**
```
=== Superblock Information ===

Magic:           _BHRfS_M ✓
Label:           TestBtrfs
FSID:            413597f2-e149-4eca-92e2-9e100f01fac5
Total Bytes:     268435456 (256.00 MB)
Bytes Used:      131072 (0.12 MB)
Usage:           0.05%
Sector Size:     4096 bytes
Node Size:       16384 bytes
...
```

### ls - List Directory Contents

List files and directories in a Btrfs filesystem.

```bash
btrfs-read ls [options] <image> [path]

Options:
  --json              Output in JSON format
  -l, --log-level     Set log level: debug, info, warn, error (default: info)
```

**Examples:**

```bash
# List root directory
btrfs-read ls tests/testdata/test.img /

# List subdirectory
btrfs-read ls tests/testdata/test.img /subdir

# Multi-level path
btrfs-read ls tests/testdata/test.img /a/b/c

# JSON output
btrfs-read ls --json tests/testdata/test.img /

# With debug logging
btrfs-read ls -l debug tests/testdata/test.img /
```

**Text Output:**
```
=== Directory Listing ===
Path: /

Type       Inode           Name
-------------------------------------------
file       257             hello.txt
file       258             readme.txt
dir        263             subdir
```

**JSON Output:**
```json
{
  "path": "/",
  "entries": [
    {
      "name": "hello.txt",
      "inode": 257,
      "type": 1,
      "is_dir": false
    },
    {
      "name": "subdir",
      "inode": 263,
      "type": 2,
      "is_dir": true
    }
  ]
}
```

### cat - Read File Content

Read and display file contents from a Btrfs filesystem.

```bash
btrfs-read cat [options] <image> <path>

Options:
  --json              Output in JSON format
  -l, --log-level     Set log level: debug, info, warn, error (default: info)
```

**Examples:**

```bash
# Read file
btrfs-read cat tests/testdata/test.img /hello.txt

# Read from nested directory
btrfs-read cat tests/testdata/test.img /dir/subdir/file.txt

# JSON output
btrfs-read cat --json tests/testdata/test.img /hello.txt

# Quiet mode (errors only)
btrfs-read cat -l error tests/testdata/test.img /hello.txt
```

**Text Output:**
```
=== Btrfs File Reader ===
Device: tests/testdata/test.img
File:   /hello.txt

✓ File read successfully (18 bytes)

=== File Content ===
Hello from Btrfs!
```

**JSON Output:**
```json
{
  "path": "/hello.txt",
  "size": 18,
  "content": "Hello from Btrfs!\n"
}
```

## Log Levels

Control the verbosity of output:

| Level | Description |
|-------|-------------|
| `debug` | Detailed debugging information (B-Tree traversal, chunk lookups) |
| `info` | Normal operation (default, clean output) |
| `warn` | Warnings only (parsing issues, non-fatal errors) |
| `error` | Errors only (fatal errors, data corruption) |

**Examples:**

```bash
# Debug mode - see internal operations
btrfs-read ls -l debug tests/testdata/test.img /

# Info mode - default, clean output
btrfs-read ls tests/testdata/test.img /

# Warn mode - only warnings
btrfs-read cat -l warn tests/testdata/test.img /file.txt

# Error mode - only errors
btrfs-read cat -l error tests/testdata/test.img /file.txt
```

**Debug Output Example:**
```
[DEBUG] 2025/12/19 11:33:41 filesystem.go:86: FS Tree root: 0x1d30000
=== Directory Listing ===
Path: /
...
```

## Output Redirection

Logs go to `stderr`, program output goes to `stdout`:

```bash
# Only see output (no logs)
btrfs-read ls tests/testdata/test.img / 2>/dev/null

# Only see logs
btrfs-read ls tests/testdata/test.img / >/dev/null

# Save JSON to file, logs to console
btrfs-read ls --json tests/testdata/test.img / >output.json

# Save both separately
btrfs-read ls tests/testdata/test.img / >output.txt 2>debug.log
```

## Troubleshooting

### Error: "no chunk mapping found"

**Cause:**
- Corrupted image file
- Unsupported RAID type
- Chunk tree loading failed

**Solution:**
1. Verify image file integrity
2. Use `mkfs.btrfs -d single -m single` for SINGLE mode
3. Check with `btrfs-read info <image>` first

### Error: "file not found"

**Cause:**
- File doesn't exist
- Path is case-sensitive
- File is in a snapshot/subvolume (not supported)

**Solution:**
1. Use `btrfs-read ls <image> /` to list available files
2. Check file path spelling and case
3. Ensure file is in the main filesystem (not a snapshot)

### Empty Output

**Cause:**
- Default log level is `info` (no debug output)

**Solution:**
Use `-l debug` to see detailed operation logs:
```bash
btrfs-read ls -l debug tests/testdata/test.img /
```

## More Information

- [Architecture Documentation](ARCHITECTURE.md) - System architecture
- [Main README](../README.md) - Project overview
- [Diagrams](../diagrams/) - Architecture diagrams
