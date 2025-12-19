# 测试和演示脚本

本目录包含用于测试和演示 Btrfs-Read 工具的脚本。

## 脚本列表

### demo.sh
基础功能演示脚本

**用途**: 展示 btrfs-read 的基本功能

**使用**:
```bash
cd /root/workspaces/btrfs_read
./scripts/demo.sh
```

**测试内容**:
- 列出根目录（文本格式）
- 列出根目录（JSON 格式）
- 列出子目录
- 读取文件（文本格式）
- 读取文件（JSON 格式）

### test_multilevel.sh
多层级目录测试脚本

**用途**: 测试多层级目录遍历功能

**使用**:
```bash
cd /root/workspaces/btrfs_read
./scripts/test_multilevel.sh
```

**测试内容**:
- 1-4 层深度的目录导航
- 深层文件读取
- 不同路径分支测试

### final_test.sh
综合测试脚本

**用途**: 全面测试所有功能

**使用**:
```bash
cd /root/workspaces/btrfs_read
./scripts/final_test.sh
```

**测试内容**:
- 构建流程
- 基础命令
- 多层级目录
- JSON 输出
- 文档引用验证

### verify_setup.sh
设置验证脚本

**用途**: 验证项目配置是否正确

**使用**:
```bash
cd /root/workspaces/btrfs_read
./scripts/verify_setup.sh
```

**验证内容**:
- 可执行文件位置
- 目录结构
- 文档引用
- 基础功能
- Makefile 配置

## 运行所有测试

```bash
cd /root/workspaces/btrfs_read

# 验证设置
./scripts/verify_setup.sh

# 运行演示
./scripts/demo.sh

# 多层级测试
./scripts/test_multilevel.sh

# 综合测试
./scripts/final_test.sh
```

## 注意事项

1. **路径**: 所有脚本应从项目根目录运行
2. **测试镜像**: 需要先创建测试镜像：
   ```bash
   sudo bash tests/create-test-image.sh
   ```
3. **构建**: 某些脚本需要先构建项目：
   ```bash
   make build
   ```

## 脚本开发

如果要添加新的测试脚本：

1. 放在 `scripts/` 目录
2. 使用 `.sh` 扩展名
3. 添加可执行权限：`chmod +x scripts/your_script.sh`
4. 使用相对路径引用项目文件（从根目录开始）
5. 更新本 README.md
