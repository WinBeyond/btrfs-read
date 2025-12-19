# Btrfs-Read 架构图

本目录包含 Btrfs-Read 项目的架构图和流程图,使用 Mermaid 格式编写。

## 图表列表

### 1. 系统架构图
**文件**: [architecture.md](architecture.md)

展示了 Btrfs-Read 的五层架构设计:
- 应用层 (Application Layer)
- 文件系统层 (Filesystem Layer)  
- B-Tree 层 (B-Tree Layer)
- 逻辑块层 (Logical Block Layer)
- 物理块层 (Physical Block Layer)

### 2. 地址映射流程
**文件**: [address-mapping.md](address-mapping.md)

说明了逻辑地址到物理地址的映射过程,包括:
- Chunk 查找
- RAID 类型处理 (SINGLE/DUP/RAID0/1/5/6/10)
- 物理地址计算

### 3. B-Tree 搜索流程
**文件**: [btree-search.md](btree-search.md)

详细展示了 B-Tree 索引搜索算法:
- 节点读取和解析
- 二分查找实现
- 内部节点和叶节点处理
- Path 路径记录

### 4. 文件读取流程
**文件**: [file-read-flow.md](file-read-flow.md)

完整的文件读取时序图,包含:
1. 路径解析 (遍历目录树)
2. Inode 信息读取
3. Extent 查找
4. 数据读取和解压
5. 校验和验证

### 5. 初始化流程
**文件**: [init-flow.md](init-flow.md)

文件系统初始化过程:
1. 设备扫描
2. Superblock 读取
3. Chunk Tree 初始化
4. FS Tree 定位
5. 缓存初始化

## 如何查看图表

### 在 GitHub 上查看
GitHub 原生支持 Mermaid 语法,直接点击 `.md` 文件即可查看渲染后的图表。

### 在本地查看

#### 方法 1: VS Code
1. 安装 "Markdown Preview Mermaid Support" 插件
2. 打开 `.md` 文件
3. 使用 `Ctrl+Shift+V` (Windows/Linux) 或 `Cmd+Shift+V` (Mac) 预览

#### 方法 2: Mermaid Live Editor
1. 访问 https://mermaid.live/
2. 复制图表代码
3. 粘贴到编辑器中查看

#### 方法 3: 使用 Mermaid CLI
```bash
# 安装 mermaid-cli
npm install -g @mermaid-js/mermaid-cli

# 生成 PNG 图片
mmdc -i architecture.md -o architecture.png

# 生成 SVG 图片
mmdc -i architecture.md -o architecture.svg
```

## 图表格式

所有图表均使用 Mermaid 格式编写,具有以下优点:
- ✅ 纯文本,易于版本控制
- ✅ GitHub 原生支持
- ✅ 可转换为多种格式 (PNG, SVG, PDF)
- ✅ 易于维护和更新

## 相关文档

- [ARCHITECTURE.md](../docs/ARCHITECTURE.md) - 详细的架构设计文档
- [PROJECT.md](../PROJECT.md) - 项目概览
- [README.md](../README.md) - 项目主页

## 贡献

如需更新图表:
1. 编辑对应的 `.md` 文件
2. 修改 Mermaid 代码块
3. 在本地预览确认无误
4. 提交 Pull Request
