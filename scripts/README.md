# Utility Scripts

This directory contains utility scripts for testing and demonstrating btrfs-read functionality.

## Scripts

### demo.sh
Basic demonstration of btrfs-read capabilities.

```bash
./scripts/demo.sh
```

**What it does:**
- Lists root directory (text and JSON)
- Lists subdirectory
- Reads files (text and JSON)

**Prerequisites:**
- btrfs-read built or installed
- Test image created (`tests/testdata/test.img`)

---

### test_multilevel.sh
Tests multi-level directory traversal functionality.

```bash
./scripts/test_multilevel.sh
```

**What it does:**
- Navigates through 4-level deep directory structure
- Tests file reading at various depths
- Demonstrates JSON output

**Test structure:**
```
/
├── level1/
│   ├── file1.txt
│   └── level2/
│       ├── file2.txt
│       └── level3/
│           ├── file3.txt
│           └── level4/
│               └── file4.txt
├── dir1/dir2/data.txt
└── a/b/c/deep.txt
```

**Prerequisites:**
- Multi-level test image (`tests/testdata/test_multilevel.img`)
- Create with: `sudo bash tests/create-multilevel-image.sh`

---

### final_test.sh
Comprehensive test suite covering all functionality.

```bash
./scripts/final_test.sh
```

**What it does:**
1. Clean build
2. Build project
3. Test basic commands (ls, cat)
4. Test JSON output
5. Test multi-level directories
6. Verify documentation

**Exit codes:**
- 0: All tests passed
- 1: One or more tests failed

---

### verify_setup.sh
Verifies project setup and configuration.

```bash
./scripts/verify_setup.sh
```

**What it does:**
- Checks executable location (`build/btrfs-read`)
- Ensures root directory is clean (no stray executables)
- Tests basic command execution
- Verifies documentation references
- Tests Makefile

---

## Running All Tests

```bash
# Run all scripts in sequence
./scripts/verify_setup.sh
./scripts/demo.sh
./scripts/test_multilevel.sh
./scripts/final_test.sh
```

Or use the Makefile:
```bash
make test-cli
```

## Creating Test Images

Before running scripts, create test images:

```bash
# Basic test image
sudo bash tests/create-test-image.sh

# Multi-level test image
sudo bash tests/create-multilevel-image.sh
```

## Script Requirements

All scripts assume:
- Working directory is project root
- btrfs-read is built in `build/btrfs-read` or installed globally
- Test images exist in `tests/testdata/`

## Troubleshooting

### Permission denied
Scripts need execute permission:
```bash
chmod +x scripts/*.sh
```

### Test image not found
Create test images first:
```bash
sudo bash tests/create-test-image.sh
```

### btrfs-read command not found
Either:
1. Build locally: `make build`
2. Or install: `go install github.com/WinBeyond/btrfs-read/cmd/btrfs-read@latest`

## Adding New Scripts

When adding new scripts:
1. Add execute permission: `chmod +x scripts/newscript.sh`
2. Use `#!/bin/bash` shebang
3. Run from project root
4. Update this README
5. Add to `.gitignore` if temporary

## Example Output

### demo.sh
```
=== Btrfs Reader CLI Demo ===

1. List root directory (text output):
=== Directory Listing ===
Path: /

Type       Inode           Name
-------------------------------------------
file       257             hello.txt
...
```

### final_test.sh
```
==========================================
  Final Comprehensive Test
==========================================

1. Clean build...
   ✓ Cleaned

2. Build project...
   ✓ Build successful
...

==========================================
  ✓ All Tests Passed Successfully!
==========================================
```
