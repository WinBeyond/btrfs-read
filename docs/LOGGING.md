# 日志系统说明

## 概述

Btrfs-Read 使用分级日志系统，支持 4 个日志级别。默认级别为 `info`。

## 日志级别

### 1. DEBUG
最详细的日志，用于开发和调试。

**包含内容**:
- FS Tree root 地址
- B-Tree 搜索细节
- DIR_ITEM 查找过程
- 内部函数调用跟踪
- 数据结构解析细节

**使用场景**:
- 调试文件查找问题
- 追踪 B-Tree 遍历
- 开发新功能
- 理解文件系统内部工作原理

**示例**:
```bash
./build/btrfs-read ls -l debug tests/testdata/test.img /
```

**输出示例**:
```
[DEBUG] 2025/12/19 11:33:41 filesystem.go:86: FS Tree root: 0x1d30000
```

### 2. INFO (默认)
正常操作的关键信息。

**包含内容**:
- （当前没有 info 级别日志）
- 未来可能添加文件系统操作统计信息

**使用场景**:
- 日常使用（默认级别）
- 干净的输出，无调试信息

**示例**:
```bash
./build/btrfs-read ls tests/testdata/test.img /
# 或
./build/btrfs-read ls -l info tests/testdata/test.img /
```

### 3. WARN
警告信息，操作可以继续但可能有问题。

**包含内容**:
- 系统 chunk 数组为空
- 无法解析某些 chunk（可能是不支持的 RAID 类型）
- 数据格式异常但可以容错

**使用场景**:
- 生产环境
- 只关注潜在问题
- 减少日志输出

**示例**:
```bash
./build/btrfs-read ls -l warn tests/testdata/test.img /
```

### 4. ERROR
错误信息，操作无法继续。

**包含内容**:
- 数据超出边界
- 严重的格式错误
- 致命性失败

**使用场景**:
- 最小化日志输出
- 只记录错误

**示例**:
```bash
./build/btrfs-read ls -l error tests/testdata/test.img /
```

## 日志分类

### Debug 级别日志
位于 `pkg/fs/filesystem.go`:
```go
logger.Debug("FS Tree root: 0x%x", fsTreeRoot)
logger.Debug("DIR_ITEM search failed: %v", err)
logger.Debug("Failed to get DIR_ITEM: %v", err)
```

### Info 级别日志
当前没有 info 级别的日志。

### Warn 级别日志
位于 `pkg/chunk/loader.go`:
```go
logger.Warn("Failed to parse chunk at offset 0x%x: %v", item.Key.Offset, err)
```

位于 `pkg/chunk/manager.go`:
```go
logger.Warn("System chunk array size is 0")
```

### Error 级别日志
位于 `pkg/chunk/manager.go`:
```go
logger.Error("Stripe data exceeds array size: need %d bytes, only %d remaining", ...)
```

## 命令行使用

### 全局选项
```bash
--log-level <level>   # 长格式
-l <level>            # 短格式
```

### 所有命令支持
```bash
# ls 命令
./build/btrfs-read ls -l debug <image> [path]
./build/btrfs-read ls --log-level info <image> [path]

# cat 命令
./build/btrfs-read cat -l warn <image> <path>
./build/btrfs-read cat --log-level error <image> <path>

# info 命令（暂不支持日志选项）
./build/btrfs-read info <image>
```

### 结合 JSON 输出
```bash
# 日志输出到 stderr，JSON 输出到 stdout
./build/btrfs-read ls --json -l debug tests/testdata/test.img / 2>debug.log >output.json
```

## 日志输出位置

- **日志**: 输出到 `stderr`
- **程序输出**: 输出到 `stdout`

这样可以方便地分离日志和实际数据：

```bash
# 只看日志
./build/btrfs-read ls tests/testdata/test.img / 2>&1 >/dev/null

# 只看输出
./build/btrfs-read ls tests/testdata/test.img / 2>/dev/null

# JSON 输出到文件，日志到控制台
./build/btrfs-read ls --json tests/testdata/test.img / >output.json
```

## 日志格式

### Debug/Error
包含文件名和行号:
```
[DEBUG] 2025/12/19 11:29:45 filesystem.go:354: DIR_ITEM search failed: not found
[ERROR] 2025/12/19 11:29:45 manager.go:133: Stripe data exceeds array size
```

### Info/Warn
不包含文件名:
```
[INFO]  2025/12/19 11:29:45 FS Tree root: 0x1d30000
[WARN]  2025/12/19 11:29:45 System chunk array size is 0
```

## 性能考虑

- 日志检查是在运行时进行的
- 如果日志级别设为 `error`，`debug`/`info`/`warn` 日志不会被处理
- 对性能影响极小

## 编程接口

在代码中使用日志:

```go
import "github.com/WinBeyond/btrfs-read/pkg/logger"

// Debug 日志
logger.Debug("Variable value: %v", value)

// Info 日志
logger.Info("Operation completed successfully")

// Warn 日志
logger.Warn("Potential issue: %s", issue)

// Error 日志
logger.Error("Fatal error: %v", err)
```

## 默认行为

如果不指定 `--log-level` 或 `-l`，默认使用 `info` 级别：
- 显示重要的操作信息
- 不显示调试细节
- 不会过于冗长

## 最佳实践

1. **开发调试**: 使用 `-l debug`
2. **日常使用**: 使用默认 (info) 或 `-l info`
3. **生产环境**: 使用 `-l warn` 或 `-l error`
4. **性能测试**: 使用 `-l error` 最小化开销
5. **问题排查**: 使用 `-l debug` 并保存日志文件

示例：
```bash
# 开发调试
./build/btrfs-read ls -l debug tests/testdata/test.img / 2>debug.log

# 生产使用
./build/btrfs-read ls -l warn tests/testdata/test.img / 2>>production.log

# JSON API 调用（只要数据，不要日志）
./build/btrfs-read ls --json -l error tests/testdata/test.img / 2>/dev/null
```
