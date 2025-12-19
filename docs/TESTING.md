# Btrfs Read Service - 测试指南

## 目录

1. [快速开始](#快速开始)
2. [创建测试镜像](#创建测试镜像)
3. [运行测试](#运行测试)
4. [测试覆盖](#测试覆盖)
5. [手动验证](#手动验证)

---

## 快速开始

```bash
# 1. 创建测试镜像（需要 root 权限）
sudo ./tests/create-test-image.sh

# 2. 运行所有测试
make test

# 3. 运行 CLI 工具测试
make build
./build/btrfs-read tests/testdata/test.img
```

---

## 创建测试镜像

### 自动创建（推荐）

使用提供的脚本自动创建测试镜像：

```bash
# 需要 root 权限
sudo ./tests/create-test-image.sh
```

脚本会：
- ✅ 创建 100MB 的镜像文件
- ✅ 格式化为 Btrfs 文件系统（标签：TestBtrfs）
- ✅ 创建测试目录结构
- ✅ 创建各种测试文件（文本、二进制、符号链接）
- ✅ 验证魔数
- ✅ 显示文件系统信息

### 测试数据结构

创建的测试镜像包含以下结构：

```
/
├── hello.txt              # 简单文本文件
├── test.txt               # 测试描述文件
├── small.bin              # 1KB 随机数据
├── medium.bin             # 64KB 随机数据
├── large.bin              # 512KB 随机数据
├── link.txt -> hello.txt  # 符号链接
├── file with spaces.txt   # 特殊文件名
│
├── home/
│   └── user/
│       └── data.txt       # 用户数据
│
├── etc/
│   └── config.conf        # 配置文件
│
├── var/
│   └── log/
│       └── test.log       # 日志文件
│
└── deep/
    └── nested/
        └── directory/
            └── structure/
                └── file.txt  # 深层文件
```

### 手动创建（可选）

如果自动脚本失败，可以手动创建：

```bash
# 1. 创建镜像文件
truncate -s 100M tests/testdata/test.img

# 2. 格式化
mkfs.btrfs -f -L TestBtrfs tests/testdata/test.img

# 3. 挂载
sudo mkdir -p /mnt/test
sudo mount -o loop tests/testdata/test.img /mnt/test

# 4. 创建测试数据
sudo sh -c 'echo "Hello Btrfs!" > /mnt/test/hello.txt'
sudo mkdir -p /mnt/test/home/user
sudo sh -c 'echo "test data" > /mnt/test/home/user/data.txt'

# 5. 卸载
sudo umount /mnt/test
```

---

## 运行测试

### 单元测试

测试单个包：

```bash
# 测试 ondisk 包
go test ./pkg/ondisk/... -v

# 测试 errors 包
go test ./pkg/errors/... -v

# 测试所有包
go test ./pkg/... -v
```

### 集成测试

集成测试需要测试镜像：

```bash
# 确保测试镜像存在
ls -lh tests/testdata/test.img

# 运行集成测试
go test ./tests/integration/... -v

# 带详细输出（显示 Superblock 信息）
go test ./tests/integration/... -v -test.v
```

### 基准测试

```bash
# 运行所有基准测试
make bench

# 或者
go test ./... -bench=. -benchmem

# 特定基准测试
go test ./pkg/ondisk -bench=BenchmarkSuperblock -benchmem
go test ./tests/integration -bench=BenchmarkReadSuperblock -benchmem
```

### 测试覆盖率

```bash
# 生成覆盖率报告
make coverage

# 这会：
# 1. 运行所有测试
# 2. 生成 coverage.out
# 3. 生成 coverage.html
# 4. 可以在浏览器中查看

# 查看覆盖率
go tool cover -func=coverage.out

# 在浏览器中查看
open coverage.html  # macOS
xdg-open coverage.html  # Linux
```

---

## 测试覆盖

### 已实现的测试

#### pkg/ondisk/superblock_test.go ✅

**TestSuperblockUnmarshal**
- 测试基本解析功能
- 验证魔数
- 验证 IsValid() 方法

**TestSuperblockInvalidMagic**
- 测试错误魔数的处理
- 验证错误返回

**TestSuperblockTooShort**
- 测试缓冲区太短的情况
- 边界条件测试

**TestSuperblockGetLabel**
- 测试标签提取
- 测试空标签
- 测试带 null 终止符的标签

**TestSuperblockFields**
- 测试所有字段的正确解析
- 验证字节序转换
- 验证数值正确性

**BenchmarkSuperblockUnmarshal**
- 性能基准测试
- 测量解析速度

#### tests/integration/read_superblock_test.go ✅

**TestReadSuperblockFromImage**
- 从真实镜像读取 Superblock
- 验证魔数
- 验证所有关键字段
- 子测试：BasicFields, Label, TreeRoots, DeviceInfo

**TestReadBackupSuperblock**
- 测试读取备份 Superblock
- 测试多个备份位置（主、备份1、备份2）
- 验证生成号

**BenchmarkReadSuperblock**
- 测量从镜像读取和解析的性能

### 待实现的测试

**Phase 2: Chunk 层测试**
- [ ] pkg/chunk/chunk_test.go
- [ ] pkg/chunk/manager_test.go
- [ ] tests/integration/chunk_mapping_test.go

**Phase 3: B-Tree 层测试**
- [ ] pkg/btree/key_test.go
- [ ] pkg/btree/node_test.go
- [ ] pkg/btree/search_test.go
- [ ] tests/integration/btree_search_test.go

**Phase 4: 文件系统层测试**
- [ ] pkg/fs/inode_test.go
- [ ] pkg/fs/dir_test.go
- [ ] pkg/fs/file_test.go
- [ ] tests/integration/read_file_test.go
- [ ] tests/integration/read_dir_test.go

---

## 手动验证

### 验证 Superblock

使用 `xxd` 查看原始数据：

```bash
# 查看主 Superblock（偏移 0x10000 = 65536）
xxd -s 65536 -l 4096 tests/testdata/test.img | less

# 查找魔数（应该在偏移 +64 字节，即 65600）
xxd -s 65600 -l 8 tests/testdata/test.img
# 应该看到: _BHRfS_M

# 使用 btrfs 工具查看
sudo btrfs inspect-internal dump-super tests/testdata/test.img
```

### 使用 CLI 工具

```bash
# 编译 CLI 工具
make build

# 查看 Superblock 信息
./build/btrfs-read tests/testdata/test.img
```

期望输出：

```
=== Btrfs CLI Tool ===
Reading device: tests/testdata/test.img

✓ Successfully read superblock data

✓ Successfully parsed superblock

=== Superblock Information ===

Magic:           _BHRfS_M ✓
Label:           TestBtrfs
FSID:            xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
...
```

### 挂载验证

挂载测试镜像并验证内容：

```bash
# 挂载
sudo mkdir -p /mnt/test
sudo mount -o loop tests/testdata/test.img /mnt/test

# 查看内容
ls -la /mnt/test
cat /mnt/test/hello.txt

# 验证文件树
tree /mnt/test

# 卸载
sudo umount /mnt/test
```

### 使用 btrfs-progs 工具验证

```bash
# 显示文件系统信息
sudo btrfs filesystem show tests/testdata/test.img

# 显示 Superblock
sudo btrfs inspect-internal dump-super tests/testdata/test.img

# 检查文件系统
sudo btrfs check tests/testdata/test.img

# 显示树结构（需要挂载）
sudo btrfs inspect-internal dump-tree tests/testdata/test.img
```

---

## 持续集成

### GitHub Actions 配置（示例）

创建 `.github/workflows/test.yml`：

```yaml
name: Test

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Install btrfs-progs
        run: sudo apt-get install -y btrfs-progs
      
      - name: Create test image
        run: sudo ./tests/create-test-image.sh
      
      - name: Run tests
        run: make test
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out
```

---

## 故障排除

### 问题：测试镜像创建失败

**错误**: `mkfs.btrfs: command not found`

**解决**:
```bash
# Ubuntu/Debian
sudo apt install btrfs-progs

# CentOS/RHEL
sudo yum install btrfs-progs

# Arch Linux
sudo pacman -S btrfs-progs
```

### 问题：权限不足

**错误**: `mount: permission denied`

**解决**: 使用 `sudo` 运行测试镜像创建脚本

### 问题：测试跳过

**信息**: `Test image not found. Run 'make create-test-image' first.`

**解决**:
```bash
sudo make create-test-image
# 或
sudo ./tests/create-test-image.sh
```

### 问题：无法挂载镜像

**错误**: `mount: ... already mounted`

**解决**:
```bash
# 查找挂载点
mount | grep test.img

# 卸载
sudo umount /mnt/test

# 或强制卸载
sudo umount -f /mnt/test
```

---

## 测试最佳实践

1. **总是先创建测试镜像**
   ```bash
   sudo ./tests/create-test-image.sh
   ```

2. **使用 -v 标志查看详细输出**
   ```bash
   go test ./tests/integration/... -v
   ```

3. **运行基准测试了解性能**
   ```bash
   go test ./... -bench=. -benchmem
   ```

4. **检查覆盖率**
   ```bash
   make coverage
   ```

5. **提交前运行所有测试**
   ```bash
   make test
   ```

---

## 下一步测试计划

### Phase 1 完成后（Superblock ✅）

**Phase 2: Chunk 层测试**
- 创建包含不同 RAID 配置的测试镜像
- 测试地址映射算法
- 测试红黑树查找性能

**Phase 3: B-Tree 层测试**
- 创建包含深层目录的测试镜像
- 测试递归搜索
- 测试二分查找边界情况

**Phase 4: 文件系统层测试**
- 测试路径解析
- 测试文件读取（各种大小）
- 测试目录遍历
- 测试符号链接
- 测试压缩文件

---

**更新日期**: 2025-12-18  
**测试覆盖率目标**: 80%+
