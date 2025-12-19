# Btrfs Read Service - 快速开始指南

## 🚀 5 分钟快速体验

本指南将帮你快速创建测试镜像并验证 Btrfs 读取功能。

---

## 前置要求

确保系统已安装以下工具：

```bash
# Ubuntu/Debian
sudo apt install -y golang-go btrfs-progs

# CentOS/RHEL
sudo yum install -y golang btrfs-progs

# Arch Linux
sudo pacman -S go btrfs-progs

# macOS (btrfs-progs 不可用，需要 Linux 环境)
brew install go
```

验证安装：

```bash
go version          # 应该显示 Go 1.21 或更高
mkfs.btrfs --version  # 应该显示 btrfs-progs 版本
```

---

## 步骤 1: 克隆项目

```bash
git clone https://github.com/yourname/btrfs-read.git
cd btrfs-read
```

---

## 步骤 2: 下载依赖

```bash
make deps
```

这会下载所有 Go 依赖包。

---

## 步骤 3: 创建测试镜像

```bash
# 需要 root 权限
sudo make create-test-image
```

**这会做什么？**
- 创建 100MB 的镜像文件 (`tests/testdata/test.img`)
- 格式化为 Btrfs 文件系统
- 创建测试目录和文件
- 验证镜像有效性

**预期输出：**

```
=== Btrfs 测试镜像创建工具 ===
✓ 工具检查通过
✓ 镜像文件创建成功
✓ Btrfs 格式化成功
✓ 挂载成功
✓ 测试数据创建成功
✓ 卸载成功

=== 测试镜像创建完成 ===
镜像位置: /path/to/tests/testdata/test.img
镜像大小: 100M
```

---

## 步骤 4: 运行单元测试

```bash
make test-unit
```

**测试内容：**
- Superblock 解析
- 数据结构验证
- 错误处理

**预期输出：**

```
Running unit tests...
=== RUN   TestSuperblockUnmarshal
--- PASS: TestSuperblockUnmarshal (0.00s)
=== RUN   TestSuperblockInvalidMagic
--- PASS: TestSuperblockInvalidMagic (0.00s)
...
PASS
ok      github.com/yourname/btrfs-read/pkg/ondisk
```

---

## 步骤 5: 运行集成测试

```bash
make test-integration
```

**测试内容：**
- 从真实镜像读取 Superblock
- 验证所有字段
- 测试备份 Superblock

**预期输出：**

```
Running integration tests...
=== RUN   TestReadSuperblockFromImage
    read_superblock_test.go:25: Reading test image: .../test.img
=== RUN   TestReadSuperblockFromImage/BasicFields
=== RUN   TestReadSuperblockFromImage/Label
    read_superblock_test.go:68: Label: TestBtrfs
=== RUN   TestReadSuperblockFromImage/TreeRoots
    read_superblock_test.go:79: Root tree: 0x1234000
    read_superblock_test.go:80: Chunk root: 0x5678000
--- PASS: TestReadSuperblockFromImage (0.01s)
PASS
```

---

## 步骤 6: 构建并测试 CLI 工具

```bash
# 构建
make build

# 运行 CLI 工具
./build/btrfs-read tests/testdata/test.img
```

**预期输出：**

```
=== Btrfs CLI Tool ===
Reading device: tests/testdata/test.img

✓ Successfully read superblock data

✓ Successfully parsed superblock

=== Superblock Information ===

Magic:           _BHRfS_M ✓
Label:           TestBtrfs
FSID:            12345678-1234-1234-1234-123456789abc
Total Bytes:     104857600 (100.00 MB)
Bytes Used:      16777216 (16.00 MB)
Usage:           16.00%
Sector Size:     4096 bytes
Node Size:       16384 bytes

--- Tree Roots ---
Root Tree:       0x1234000 (level 0)
Chunk Tree:      0x5678000 (level 0)

--- Device Information ---
Num Devices:     1
Device ID:       1
...
```

---

## 步骤 7: 手动验证镜像

### 使用 xxd 查看原始数据

```bash
# 查看魔数（应该在偏移 65600）
xxd -s 65600 -l 8 tests/testdata/test.img
```

**预期输出：**
```
00010040: 5f42 4852 6653 5f4d                      _BHRfS_M
```

### 使用 btrfs 工具

```bash
# 显示 Superblock
sudo btrfs inspect-internal dump-super tests/testdata/test.img

# 检查文件系统
sudo btrfs check tests/testdata/test.img
```

### 挂载并查看内容

```bash
# 挂载
sudo mkdir -p /mnt/test
sudo mount -o loop tests/testdata/test.img /mnt/test

# 查看内容
ls -la /mnt/test
cat /mnt/test/hello.txt
tree /mnt/test

# 卸载
sudo umount /mnt/test
```

---

## 步骤 8: 查看覆盖率

```bash
make coverage
```

这会生成 `coverage.html`，在浏览器中打开查看详细覆盖率。

---

## 常见问题

### Q1: 创建测试镜像时出错

**错误**: `mkfs.btrfs: command not found`

**解决**:
```bash
sudo apt install btrfs-progs  # Ubuntu/Debian
sudo yum install btrfs-progs  # CentOS/RHEL
```

### Q2: 集成测试跳过

**信息**: `Test image not found. Run 'make create-test-image' first.`

**解决**:
```bash
sudo make create-test-image
```

### Q3: 权限不足

**错误**: `permission denied`

**解决**: 使用 `sudo` 运行需要 root 权限的命令

### Q4: Go 版本太低

**错误**: `go: module requires Go 1.21`

**解决**:
```bash
# 下载并安装最新 Go
wget https://go.dev/dl/go1.21.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

---

## 下一步

恭喜！你已经成功：
- ✅ 创建了 Btrfs 测试镜像
- ✅ 运行了单元测试和集成测试
- ✅ 验证了 Superblock 读取功能
- ✅ 使用了 CLI 工具

**继续开发：**

1. **阅读架构文档**
   ```bash
   cat ARCHITECTURE.md
   ```

2. **查看测试指南**
   ```bash
   cat tests/TESTING.md
   ```

3. **开始实现 Phase 2（Chunk 层）**
   - 编辑 `pkg/chunk/chunk.go`
   - 编辑 `pkg/chunk/manager.go`
   - 添加测试

4. **参考资料**
   - [Btrfs Wiki](https://btrfs.wiki.kernel.org/)
   - [Btrfs On-disk Format](https://btrfs.wiki.kernel.org/index.php/On-disk_Format)
   - [btrfs-fuse 源码](https://github.com/adam900710/btrfs-fuse)

---

## 完整命令汇总

```bash
# 设置项目
git clone https://github.com/yourname/btrfs-read.git
cd btrfs-read
make deps

# 创建测试环境
sudo make create-test-image

# 运行测试
make test-unit
make test-integration
make test

# 构建和运行
make build
./build/btrfs-read tests/testdata/test.img

# 代码质量
make fmt
make vet
make coverage

# 基准测试
make bench

# 清理
make clean
```

---

**祝你开发顺利！** 🎉

如有问题，请查看 [TESTING.md](TESTING.md) 或提交 Issue。
