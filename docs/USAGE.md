# Btrfs 读取服务使用指南

## 快速开始

### 1. 编译项目

```bash
cd /root/workspaces/btrfs_read
go build -o build/btrfs-read ./cmd/btrfs-read
```

### 2. 查看文件系统信息

```bash
./build/btrfs-read info <btrfs_image>
```

示例：
```bash
./build/btrfs-read info tests/testdata/test.img
```

输出：
```
=== Superblock Information ===

Magic:           _BHRfS_M ✓
Label:           TestBtrfs
FSID:            413597f2-e149-4eca-92e2-9e100f01fac5
Total Bytes:     268435456 (256.00 MB)
Bytes Used:      131072 (0.12 MB)
Usage:           0.05%
Sector Size:     4096 bytes
Node Size:       16384 bytes
...
```

### 3. 读取文件内容

```bash
./build/btrfs-read cat <btrfs_image> <file_path>
```

示例：
```bash
./build/btrfs-read cat tests/testdata/test.img /hello.txt
```

输出：
```
=== Btrfs File Reader ===
Device: tests/testdata/test.img
File:   /hello.txt

✓ FS Tree root: 0x1d08000
✓ Filesystem opened successfully
✓ File read successfully (18 bytes)

=== File Content ===
Hello from Btrfs!
```

## 创建测试镜像

如果你想创建自己的测试镜像：

```bash
# 创建 256MB 镜像文件
dd if=/dev/zero of=my_btrfs.img bs=1M count=256

# 格式化为 Btrfs
mkfs.btrfs -L MyBtrfs my_btrfs.img

# 挂载
sudo mkdir -p /mnt/btrfs_test
sudo mount -o loop my_btrfs.img /mnt/btrfs_test

# 添加文件
echo "Hello World" | sudo tee /mnt/btrfs_test/hello.txt

# 卸载
sudo umount /mnt/btrfs_test

# 现在可以用我们的工具读取
./build/btrfs-read cat my_btrfs.img /hello.txt
```

## 限制

当前版本的限制：

1. **仅支持根目录**：只能读取根目录下的文件，如 `/hello.txt`，不支持 `/subdir/file.txt`
2. **简单 Chunk 类型**：仅支持 SINGLE 和 DUP，不支持 RAID0/1/5/6/10
3. **无压缩支持**：不支持读取压缩文件
4. **无校验**：不验证数据校验和

## 代码示例

如果你想在自己的 Go 项目中使用：

```go
package main

import (
    "fmt"
    "github.com/yourname/btrfs-read/pkg/fs"
)

func main() {
    // 打开文件系统
    filesystem, err := fs.Open("my_btrfs.img")
    if err != nil {
        panic(err)
    }
    defer filesystem.Close()
    
    // 读取文件
    data, err := filesystem.ReadFile("/hello.txt")
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("File content: %s\n", string(data))
}
```

## 故障排查

### 错误：no chunk mapping found

这通常意味着：
- 镜像文件损坏
- 使用了不支持的 RAID 类型
- Chunk tree 加载失败

解决方法：
1. 检查镜像文件是否完整
2. 使用 `mkfs.btrfs -d single -m single` 创建 SINGLE 模式的文件系统

### 错误：file not found

可能原因：
- 文件不在根目录（当前不支持子目录）
- 文件名大小写不匹配
- 文件确实不存在

解决方法：
1. 确保文件在根目录
2. 检查文件名拼写

## 更多信息

- 查看 `COMPLETION_SUMMARY.md` 了解实现细节
- 查看 `ARCHITECTURE.md` 了解系统架构
- 参考源码中的注释了解具体实现
