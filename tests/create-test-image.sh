#!/bin/bash

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 配置
IMAGE_DIR="$(dirname "$0")/testdata"
IMAGE_FILE="$IMAGE_DIR/test.img"
IMAGE_SIZE="100M"
MOUNT_POINT="/tmp/btrfs-test-$$"

echo -e "${GREEN}=== Btrfs 测试镜像创建工具 ===${NC}"

# 检查是否有 root 权限
if [ "$EUID" -ne 0 ]; then 
    echo -e "${RED}错误: 需要 root 权限运行此脚本${NC}"
    echo "请使用: sudo $0"
    exit 1
fi

# 检查必要的工具
echo -e "${YELLOW}检查必要工具...${NC}"
for cmd in truncate mkfs.btrfs mount umount; do
    if ! command -v $cmd &> /dev/null; then
        echo -e "${RED}错误: 未找到命令 '$cmd'${NC}"
        echo "请安装 btrfs-progs: sudo apt install btrfs-progs"
        exit 1
    fi
done
echo -e "${GREEN}✓ 工具检查通过${NC}"

# 创建目录
mkdir -p "$IMAGE_DIR"

# 删除旧镜像（如果存在）
if [ -f "$IMAGE_FILE" ]; then
    echo -e "${YELLOW}删除旧镜像...${NC}"
    rm -f "$IMAGE_FILE"
fi

# 1. 创建空白镜像文件
echo -e "${YELLOW}创建 ${IMAGE_SIZE} 镜像文件...${NC}"
truncate -s $IMAGE_SIZE "$IMAGE_FILE"
echo -e "${GREEN}✓ 镜像文件创建成功${NC}"

# 2. 格式化为 btrfs
echo -e "${YELLOW}格式化为 Btrfs 文件系统...${NC}"
mkfs.btrfs -f -L "TestBtrfs" "$IMAGE_FILE" > /dev/null 2>&1
echo -e "${GREEN}✓ Btrfs 格式化成功${NC}"

# 3. 挂载镜像
echo -e "${YELLOW}挂载镜像到 ${MOUNT_POINT}...${NC}"
mkdir -p "$MOUNT_POINT"
mount -o loop "$IMAGE_FILE" "$MOUNT_POINT"
echo -e "${GREEN}✓ 挂载成功${NC}"

# 4. 创建测试文件和目录
echo -e "${YELLOW}创建测试数据...${NC}"

# 简单文本文件
echo "Hello Btrfs!" > "$MOUNT_POINT/hello.txt"
echo "This is a test file for Btrfs read service." > "$MOUNT_POINT/test.txt"

# 创建目录结构
mkdir -p "$MOUNT_POINT/home/user"
mkdir -p "$MOUNT_POINT/etc"
mkdir -p "$MOUNT_POINT/var/log"

# 在目录中创建文件
echo "user data" > "$MOUNT_POINT/home/user/data.txt"
echo "config file" > "$MOUNT_POINT/etc/config.conf"
echo "log entry 1" > "$MOUNT_POINT/var/log/test.log"

# 创建不同大小的文件
dd if=/dev/urandom of="$MOUNT_POINT/small.bin" bs=1K count=1 2>/dev/null
dd if=/dev/urandom of="$MOUNT_POINT/medium.bin" bs=1K count=64 2>/dev/null
dd if=/dev/urandom of="$MOUNT_POINT/large.bin" bs=1K count=512 2>/dev/null

# 创建符号链接
ln -s hello.txt "$MOUNT_POINT/link.txt"

# 创建多级目录
mkdir -p "$MOUNT_POINT/deep/nested/directory/structure"
echo "deep file" > "$MOUNT_POINT/deep/nested/directory/structure/file.txt"

# 创建一些特殊文件名
touch "$MOUNT_POINT/file with spaces.txt"
echo "special" > "$MOUNT_POINT/file with spaces.txt"

echo -e "${GREEN}✓ 测试数据创建成功${NC}"

# 5. 显示文件系统信息
echo -e "${YELLOW}文件系统信息:${NC}"
btrfs filesystem show "$MOUNT_POINT" 2>/dev/null || true
echo ""

echo -e "${YELLOW}目录树结构:${NC}"
tree "$MOUNT_POINT" 2>/dev/null || find "$MOUNT_POINT" -type f -o -type d | sort

# 6. 卸载
echo -e "${YELLOW}卸载镜像...${NC}"
umount "$MOUNT_POINT"
rmdir "$MOUNT_POINT"
echo -e "${GREEN}✓ 卸载成功${NC}"

# 7. 显示镜像信息
echo ""
echo -e "${GREEN}=== 测试镜像创建完成 ===${NC}"
echo -e "镜像位置: ${YELLOW}$IMAGE_FILE${NC}"
echo -e "镜像大小: ${YELLOW}$(du -h "$IMAGE_FILE" | cut -f1)${NC}"
echo ""

# 8. 读取 Superblock 进行验证
echo -e "${YELLOW}验证 Superblock (读取前 512 字节)...${NC}"
echo ""
echo "魔数位置应该在偏移 0x10040 (65600):"
xxd -s 65600 -l 8 "$IMAGE_FILE" | grep -E "_BH.*fS_M" && echo -e "${GREEN}✓ 魔数验证成功${NC}" || echo -e "${RED}✗ 魔数验证失败${NC}"

echo ""
echo -e "${GREEN}可以使用以下命令测试读取:${NC}"
echo -e "  ${YELLOW}# 查看 superblock 十六进制${NC}"
echo -e "  xxd -s 65536 -l 4096 $IMAGE_FILE | less"
echo ""
echo -e "  ${YELLOW}# 重新挂载测试${NC}"
echo -e "  sudo mount -o loop $IMAGE_FILE /mnt"
echo -e "  ls -la /mnt"
echo -e "  sudo umount /mnt"
echo ""
echo -e "  ${YELLOW}# 运行 Go 测试${NC}"
echo -e "  go test ./tests/integration/... -v"
echo ""
