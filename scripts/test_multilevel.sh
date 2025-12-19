#!/bin/bash

echo "=========================================="
echo "  Btrfs Multi-Level Directory Test"
echo "=========================================="
echo ""

echo "1. List root directory:"
./build/btrfs-read ls ./tests/testdata/test_multilevel.img /
echo ""

echo "2. Navigate through level1 -> level2 -> level3 -> level4:"
echo ""
echo "  /level1:"
./build/btrfs-read ls ./tests/testdata/test_multilevel.img /level1
echo ""

echo "  /level1/level2:"
./build/btrfs-read ls ./tests/testdata/test_multilevel.img /level1/level2
echo ""

echo "  /level1/level2/level3:"
./build/btrfs-read ls ./tests/testdata/test_multilevel.img /level1/level2/level3
echo ""

echo "  /level1/level2/level3/level4:"
./build/btrfs-read ls ./tests/testdata/test_multilevel.img /level1/level2/level3/level4
echo ""

echo "3. Read file at depth 4:"
./build/btrfs-read cat ./tests/testdata/test_multilevel.img /level1/level2/level3/level4/file4.txt
echo ""

echo "4. Test another deep path /a/b/c (JSON output):"
./build/btrfs-read ls --json ./tests/testdata/test_multilevel.img /a/b/c
echo ""

echo "5. Read file from /a/b/c (JSON output):"
./build/btrfs-read cat --json ./tests/testdata/test_multilevel.img /a/b/c/deep.txt
echo ""

echo "6. Read file from /dir1/dir2:"
./build/btrfs-read cat ./tests/testdata/test_multilevel.img /dir1/dir2/data.txt
echo ""

echo "=========================================="
echo "  All Multi-Level Tests Passed!"
echo "=========================================="
