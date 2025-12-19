# Btrfs-Read 项目概览

## 项目简介

Btrfs-Read 是一个用 Go 语言实现的 Btrfs 文件系统只读工具，可以直接读取 Btrfs 镜像文件或设备中的文件和目录，无需挂载文件系统。

## 项目结构

```
btrfs_read/
├── build/                      # 构建输出目录
│   └── btrfs-read             # 可执行文件 (3.0M)
│
├── cmd/                        # 应用程序入口
│   └── btrfs-read/            # CLI 工具源码
│       └── main.go
│
├── pkg/                        # 核心包
│   ├── btree/                 # B-Tree 实现
│   ├── chunk/                 # Chunk 管理（逻辑地址映射）
│   ├── device/                # 设备层（物理读取）
│   ├── errors/                # 错误定义
│   ├── fs/                    # 文件系统层
│   └── ondisk/                # 磁盘格式定义
│
├── tests/                      # 测试文件
│   ├── integration/           # 集成测试
│   ├── testdata/              # 测试镜像文件
│   └── create-test-image.sh   # 测试镜像创建脚本
│
├── docs/                       # 文档目录
│   ├── README.md              # 文档索引
│   ├── ARCHITECTURE.md        # 架构设计文档
│   ├── USAGE.md               # 使用说明
│   ├── QUICKSTART.md          # 快速开始
│   ├── BUILD_AND_NAMING.md    # 构建说明
│   ├── TESTING.md             # 测试文档
│   ├── TEST_REPORT.md         # 测试报告
│   └── MULTILEVEL_TEST_RESULTS.md  # 多层级测试结果
│
├── scripts/                    # 脚本工具
│   ├── README.md              # 脚本说明
│   ├── demo.sh                # 功能演示
│   ├── test_multilevel.sh     # 多层级测试
│   ├── final_test.sh          # 综合测试
│   └── verify_setup.sh        # 设置验证
│
├── diagrams/                   # 架构图 (Mermaid 格式)
│   ├── README.md              # 图表说明文档
│   ├── architecture.md        # 总体架构
│   ├── address-mapping.md     # 地址映射流程
│   ├── btree-search.md        # B-Tree 搜索
│   ├── file-read-flow.md      # 文件读取流程
│   └── init-flow.md           # 初始化流程
│
├── go.mod                      # Go 模块定义
├── Makefile                    # 构建脚本
├── README.md                   # 项目说明
└── PROJECT.md                  # 本文件 - 项目概览
```

## 核心功能

### 1. 文件系统操作
- **info**: 显示 Btrfs 超级块信息
- **ls**: 列出目录内容，支持多层级路径
- **cat**: 读取文件内容，支持任意深度路径

### 2. 输出格式
- 文本格式：易于人类阅读
- JSON 格式：便于程序解析

### 3. 技术特性
- ✅ 完整的 Btrfs 磁盘格式解析
- ✅ B-Tree 索引遍历
- ✅ Chunk 逻辑到物理地址映射
- ✅ CRC32C 校验和验证
- ✅ LRU 缓存优化
- ✅ 支持 INLINE 和 REGULAR 文件
- ✅ 多层级目录遍历

## 技术架构

### 五层架构设计

```
┌─────────────────────────────────┐
│   Application Layer (CLI)       │  ← 命令行接口
├─────────────────────────────────┤
│   Filesystem Layer (pkg/fs)     │  ← 文件系统操作
├─────────────────────────────────┤
│   B-Tree Layer (pkg/btree)      │  ← 索引搜索
├─────────────────────────────────┤
│   Chunk Layer (pkg/chunk)       │  ← 地址映射
├─────────────────────────────────┤
│   Device Layer (pkg/device)     │  ← 物理读取
└─────────────────────────────────┘
```

### 关键组件

1. **Device Layer** - 物理设备读取和缓存
2. **Chunk Manager** - 逻辑地址到物理地址的映射
3. **B-Tree Searcher** - 高效的索引查找
4. **Filesystem** - 文件和目录操作的高层接口
5. **CLI** - 用户交互界面

## 快速开始

### 构建
```bash
make build
```

### 使用
```bash
# 查看帮助
./build/btrfs-read

# 列出目录
./build/btrfs-read ls <image> /path

# 读取文件
./build/btrfs-read cat <image> /path/to/file

# JSON 输出
./build/btrfs-read ls --json <image> /
```

### 测试
```bash
# 创建测试镜像
sudo bash tests/create-test-image.sh

# 运行演示
./scripts/demo.sh

# 运行测试
make test
```

## 开发指南

### 添加新功能

1. 在相应的 `pkg/` 子包中添加实现
2. 在 `cmd/btrfs-read/main.go` 中添加 CLI 命令
3. 添加测试到 `tests/integration/`
4. 更新文档

### 代码规范

- 使用 `make fmt` 格式化代码
- 使用 `make vet` 进行静态检查
- 保持单一职责原则
- 添加必要的注释

### 测试规范

- 集成测试放在 `tests/integration/`
- 测试覆盖核心功能
- 使用真实的 Btrfs 镜像测试

## 性能特点

- **LRU 缓存**: 减少重复的磁盘读取
- **惰性加载**: 按需读取数据
- **二分查找**: B-Tree 节点高效搜索
- **直接 I/O**: 无需挂载文件系统

## 限制和注意事项

### 当前限制
- ⚠️ 只读访问（不支持写入）
- ⚠️ 不支持压缩文件
- ⚠️ 不支持加密文件
- ⚠️ 不支持快照切换
- ⚠️ 不支持子卷切换

### 已测试特性
- ✅ INLINE 类型文件（小于 4KB）
- ✅ REGULAR 类型文件
- ✅ 单设备 Btrfs
- ✅ 4 层深度目录
- ✅ CRC32C 校验

## 依赖

### 运行时依赖
- Go 1.21+

### 测试依赖
- Linux 系统（用于创建 Btrfs 镜像）
- btrfs-progs（mkfs.btrfs）
- sudo 权限（仅测试镜像创建）

## 参考资料

- [Btrfs Wiki](https://btrfs.wiki.kernel.org/)
- [Btrfs On-disk Format](https://btrfs.wiki.kernel.org/index.php/On-disk_Format)
- [btrfs-fuse 项目](https://github.com/adam900710/btrfs-fuse)

## 维护者

Btrfs-Read Project Team

## 许可证

MIT License

---

**最后更新**: 2025-12-19
**版本**: v1.0
