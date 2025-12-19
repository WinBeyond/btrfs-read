# Btrfs-Read 文档

本目录包含 Btrfs-Read 项目的所有文档。

## 文档索引

### 用户文档

- **[快速开始 (QUICKSTART.md)](QUICKSTART.md)**  
  新手快速上手指南，包含安装、构建和基本使用示例

- **[使用说明 (USAGE.md)](USAGE.md)**  
  详细的使用指南，包含所有命令和选项的说明

- **[日志系统 (LOGGING.md)](LOGGING.md)**  
  日志级别说明和使用方法

- **[构建说明 (BUILD_AND_NAMING.md)](BUILD_AND_NAMING.md)**  
  构建配置、命名规范和安装说明

### 开发文档

- **[架构设计 (ARCHITECTURE.md)](ARCHITECTURE.md)**  
  详细的技术架构文档，包含设计理念、数据流程和 Mermaid 架构图

- **[测试文档 (TESTING.md)](TESTING.md)**  
  测试策略、测试用例和测试环境配置

- **[测试报告 (TEST_REPORT.md)](TEST_REPORT.md)**  
  完整的测试覆盖和测试结果报告

- **[多层级测试 (MULTILEVEL_TEST_RESULTS.md)](MULTILEVEL_TEST_RESULTS.md)**  
  多层级目录功能的专项测试结果

### 架构图表

- **[架构图 (../diagrams/)](../diagrams/)**  
  包含系统架构、流程图等 Mermaid 格式的可视化文档

## 文档层次

```
docs/
├── README.md                      # 本文件 - 文档索引
├── QUICKSTART.md                  # 快速开始（新手）
├── USAGE.md                       # 使用说明（用户）
├── LOGGING.md                     # 日志系统（用户）
├── BUILD_AND_NAMING.md           # 构建说明（开发者）
├── ARCHITECTURE.md               # 架构设计（开发者）
├── TESTING.md                    # 测试文档（开发者）
├── TEST_REPORT.md                # 测试报告（开发者）
└── MULTILEVEL_TEST_RESULTS.md    # 专项测试（开发者）
```

## 快速链接

### 我想...

- **开始使用这个工具** → [QUICKSTART.md](QUICKSTART.md)
- **了解所有命令** → [USAGE.md](USAGE.md)
- **从源码构建** → [BUILD_AND_NAMING.md](BUILD_AND_NAMING.md)
- **理解实现原理** → [ARCHITECTURE.md](ARCHITECTURE.md)
- **查看测试覆盖** → [TEST_REPORT.md](TEST_REPORT.md)

## 外部资源

### Btrfs 规范

- [Btrfs Wiki](https://btrfs.wiki.kernel.org/)
- [On-disk Format](https://btrfs.wiki.kernel.org/index.php/On-disk_Format)
- [Btrfs 设计文档](https://btrfs.wiki.kernel.org/index.php/Btrfs_design)

### 参考项目

- [btrfs-fuse](https://github.com/adam900710/btrfs-fuse) - 本项目参考的 Btrfs 实现
- [btrfs-progs](https://github.com/kdave/btrfs-progs) - 官方 Btrfs 工具集

## 贡献文档

如果你想为文档做贡献，请遵循以下规范：

1. 使用 Markdown 格式
2. 保持文档简洁明了
3. 添加代码示例和输出示例
4. 更新本 README.md 中的索引

## 反馈

如果发现文档有任何问题或需要改进，请提交 Issue。
