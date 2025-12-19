#!/bin/bash

echo "=== Btrfs Reader CLI Demo ==="
echo ""

echo "1. List root directory (text output):"
./build/btrfs-read ls ./tests/testdata/test.img /
echo ""

echo "2. List root directory (JSON output):"
./build/btrfs-read ls --json ./tests/testdata/test.img /
echo ""

echo "3. List subdirectory /t1:"
./build/btrfs-read ls ./tests/testdata/test.img /t1
echo ""

echo "4. Read file /hello.txt (text output):"
./build/btrfs-read cat ./tests/testdata/test.img /hello.txt
echo ""

echo "5. Read file /readme.txt (JSON output):"
./build/btrfs-read cat --json ./tests/testdata/test.img /readme.txt
echo ""

echo "=== Demo Complete ==="
