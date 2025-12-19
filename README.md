# Btrfs-Read

一个用 Go 语言实现的 Btrfs 文件系统读取工具，可以直接读取 Btrfs 镜像文件或设备中的文件和目录。

## 功能特性

- ✅ 读取 Btrfs 超级块信息
- ✅ 列出目录内容（支持多层级目录）
- ✅ 读取文件内容（支持任意深度路径）
- ✅ JSON 格式输出
- ✅ 支持 INLINE 和 REGULAR 类型的文件数据
- ✅ 完整的 B-Tree 遍历
- ✅ Chunk 逻辑到物理地址映射

## 快速开始

### 安装

#### 方法 1: 使用 go install (推荐)

```bash
# 直接从 GitHub 安装最新版本
go install github.com/WinBeyond/btrfs-read/cmd/btrfs-read@latest

# 验证安装
btrfs-read --help
```

安装后，`btrfs-read` 命令将被安装到 `$GOPATH/bin` 目录（通常是 `~/go/bin`）。

确保 `$GOPATH/bin` 在你的 `PATH` 中:
```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

#### 方法 2: 从源码构建

```bash
# 克隆仓库
git clone https://github.com/WinBeyond/btrfs-read.git
cd btrfs-read

# 使用 Makefile
make build

# 或手动构建
go build -o build/btrfs-read ./cmd/btrfs-read
```

### 使用

```bash
# 显示帮助
btrfs-read

# 显示文件系统信息
btrfs-read info <image>

# 列出目录内容
btrfs-read ls <image> [path]
btrfs-read ls --json <image> /

# 读取文件
btrfs-read cat <image> <path>
btrfs-read cat --json <image> /file.txt

# 多层级目录支持
btrfs-read ls <image> /dir1/dir2/dir3
btrfs-read cat <image> /a/b/c/file.txt

# 日志级别
btrfs-read ls -l debug <image> /
btrfs-read cat --log-level warn <image> /file.txt
```

### 示例

```bash
# 创建测试镜像（需要 root 权限）
sudo bash tests/create-test-image.sh

# 列出根目录
btrfs-read ls tests/testdata/test.img /

# 读取文件
btrfs-read cat tests/testdata/test.img /hello.txt

# JSON 输出
btrfs-read ls --json tests/testdata/test.img /
```

> **注意**: 如果使用方法 2 从源码构建，命令需要加上路径前缀 `./build/btrfs-read`

## 项目结构

```
btrfs_read/
├── build/              # 构建输出
│   └── btrfs-read     # 可执行文件
├── cmd/
│   └── btrfs-read/    # CLI 工具源码
├── pkg/               # 核心包
│   ├── btree/        # B-Tree 实现
│   ├── chunk/        # Chunk 管理
│   ├── device/       # 设备操作
│   ├── errors/       # 错误定义
│   ├── fs/           # 文件系统层
│   └── ondisk/       # 磁盘格式定义
├── tests/            # 测试文件
├── docs/             # 文档
├── scripts/          # 脚本工具
└── diagrams/         # 架构图
```

## 开发

### 运行测试

```bash
# 所有测试
make test

# 单元测试
make test-unit

# 集成测试
make test-integration
```

### 代码检查

```bash
# 格式化
make fmt

# 静态检查
make vet
```

### 清理

```bash
make clean
```

## 文档

详细文档位于 `docs/` 目录：

- [架构设计](docs/ARCHITECTURE.md) - 技术架构和设计文档
- [使用说明](docs/USAGE.md) - 详细使用指南
- [快速开始](docs/QUICKSTART.md) - 快速上手教程
- [构建说明](docs/BUILD_AND_NAMING.md) - 构建和命名配置

## 测试脚本

测试脚本位于 `scripts/` 目录：

```bash
# 基础功能演示
./scripts/demo.sh

# 多层级目录测试
./scripts/test_multilevel.sh

# 综合测试
./scripts/final_test.sh

# 验证设置
./scripts/verify_setup.sh
```

## 技术栈

- **语言**: Go 1.21+
- **核心技术**:
  - Btrfs 磁盘格式解析
  - B-Tree 数据结构
  - CRC32C 校验
  - Copy-on-Write (COW) 文件系统

## 限制

- 只读访问（不支持写入）
- 不支持压缩文件
- 不支持加密文件
- 不支持快照和子卷切换

## 许可证

MIT License

## 作者

Btrfs-Read Project
