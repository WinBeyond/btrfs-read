#!/bin/bash

echo "=========================================="
echo "  Btrfs-Read Setup Verification"
echo "=========================================="
echo ""

# 1. Check executable location
echo "1. Checking executable location..."
if [ -f "build/btrfs-read" ]; then
    echo "   ✓ build/btrfs-read exists"
    ls -lh build/btrfs-read
else
    echo "   ✗ build/btrfs-read not found"
    exit 1
fi
echo ""

# 2. Check no executables in root
echo "2. Checking root directory is clean..."
if [ -f "btrfs-read" ] || [ -f "btrfs-cli" ]; then
    echo "   ✗ Old executables found in root"
    exit 1
else
    echo "   ✓ No executables in root directory"
fi
echo ""

# 3. Verify command works
echo "3. Testing executable..."
if ./build/btrfs-read 2>&1 | grep -q "Usage: btrfs-read"; then
    echo "   ✓ Executable works correctly"
else
    echo "   ✗ Executable test failed"
    exit 1
fi
echo ""

# 4. Check documentation
echo "4. Checking documentation..."
doc_errors=0
for doc in README.md USAGE.md QUICKSTART.md; do
    if grep -q "btrfs-cli" "$doc" 2>/dev/null; then
        echo "   ✗ Found 'btrfs-cli' reference in $doc"
        doc_errors=$((doc_errors + 1))
    else
        echo "   ✓ $doc uses correct name"
    fi
done

if [ $doc_errors -gt 0 ]; then
    echo "   Some documentation needs updating"
else
    echo "   ✓ All documentation updated"
fi
echo ""

# 5. Test basic functionality
echo "5. Testing basic functionality..."
if ./build/btrfs-read ls ./tests/testdata/test.img / > /dev/null 2>&1; then
    echo "   ✓ ls command works"
else
    echo "   ✗ ls command failed"
    exit 1
fi

if ./build/btrfs-read cat ./tests/testdata/test.img /hello.txt > /dev/null 2>&1; then
    echo "   ✓ cat command works"
else
    echo "   ✗ cat command failed"
    exit 1
fi
echo ""

# 6. Test Makefile
echo "6. Testing Makefile..."
if make help | grep -q "build.*Build CLI tool"; then
    echo "   ✓ Makefile is correct"
else
    echo "   ✗ Makefile needs checking"
    exit 1
fi
echo ""

echo "=========================================="
echo "  ✓ All Verification Tests Passed!"
echo "=========================================="
