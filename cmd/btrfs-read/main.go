package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/WinBeyond/btrfs-read/pkg/fs"
	"github.com/WinBeyond/btrfs-read/pkg/logger"
	"github.com/WinBeyond/btrfs-read/pkg/ondisk"
)

var (
	jsonOutput bool
	logLevel   string
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "info":
		cmdInfo()

	case "cat":
		cmdCat()

	case "ls":
		cmdLs()

	default:
		// 兼容旧版本：如果第一个参数看起来像路径，当作 info
		if len(os.Args) == 2 {
			fmt.Fprintf(os.Stderr, "Warning: implicit 'info' command is deprecated, use 'btrfs-read info <image>' instead\n\n")
			cmdInfoLegacy(os.Args[1])
		} else {
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
			printUsage()
			os.Exit(1)
		}
	}
}

func printUsage() {
	fmt.Println("Usage: btrfs-read <command> [options] [args]")
	fmt.Println("\nCommands:")
	fmt.Println("  info <image>              - Show superblock information")
	fmt.Println("  ls <image> [path]         - List directory contents")
	fmt.Println("  cat <image> <path>        - Read file content")
	fmt.Println("\nGlobal Options:")
	fmt.Println("  --log-level, -l <level>   - Set log level: debug, info, warn, error (default: info)")
	fmt.Println("\nCommand Options:")
	fmt.Println("  --json                    - Output in JSON format (for ls and cat commands)")
	fmt.Println("\nExamples:")
	fmt.Println("  btrfs-read info tests/testdata/test.img")
	fmt.Println("  btrfs-read ls tests/testdata/test.img /")
	fmt.Println("  btrfs-read ls --json tests/testdata/test.img /")
	fmt.Println("  btrfs-read cat tests/testdata/test.img /hello.txt")
	fmt.Println("  btrfs-read cat --json tests/testdata/test.img /hello.txt")
	fmt.Println("  btrfs-read -l debug ls tests/testdata/test.img /")
}

func cmdInfo() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: btrfs-read info <image>")
		os.Exit(1)
	}
	devicePath := os.Args[2]
	cmdInfoLegacy(devicePath)
}

func cmdInfoLegacy(devicePath string) {
	fmt.Printf("=== Btrfs CLI Tool ===\n")
	fmt.Printf("Reading device: %s\n\n", devicePath)

	// 打开设备
	file, err := os.Open(devicePath)
	if err != nil {
		fmt.Printf("Error opening device: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// 读取 Superblock
	const superblockOffset = 0x10000
	const superblockSize = 4096

	buf := make([]byte, superblockSize)
	n, err := file.ReadAt(buf, superblockOffset)
	if err != nil {
		fmt.Printf("Error reading superblock: %v\n", err)
		os.Exit(1)
	}
	if n != superblockSize {
		fmt.Printf("Error: read %d bytes, expected %d\n", n, superblockSize)
		os.Exit(1)
	}

	fmt.Println("✓ Successfully read superblock data")
	fmt.Println()

	// 解析 Superblock
	sb := &ondisk.Superblock{}
	if err := sb.Unmarshal(buf); err != nil {
		fmt.Printf("Error parsing superblock: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Successfully parsed superblock")
	fmt.Println()

	// 显示 Superblock 信息
	printSuperblockInfo(sb)
}

func cmdCat() {
	flagSet := flag.NewFlagSet("cat", flag.ExitOnError)
	flagSet.BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	flagSet.StringVar(&logLevel, "log-level", "info", "Log level")
	flagSet.StringVar(&logLevel, "l", "info", "Log level (shorthand)")
	flagSet.Parse(os.Args[2:])

	// 设置日志级别
	if err := logger.SetLevelFromString(logLevel); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if flagSet.NArg() < 2 {
		fmt.Fprintln(os.Stderr, "Usage: btrfs-read cat [--json] [-l level] <image> <path>")
		os.Exit(1)
	}

	devicePath := flagSet.Arg(0)
	filePath := flagSet.Arg(1)

	// 打开文件系统
	filesystem, err := fs.Open(devicePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening filesystem: %v\n", err)
		os.Exit(1)
	}
	defer filesystem.Close()

	// 读取文件
	data, err := filesystem.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	if jsonOutput {
		// JSON 输出格式
		output := map[string]interface{}{
			"path":    filePath,
			"size":    len(data),
			"content": string(data),
		}
		jsonData, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(jsonData))
	} else {
		// 普通文本输出
		fmt.Printf("=== Btrfs File Reader ===\n")
		fmt.Printf("Device: %s\n", devicePath)
		fmt.Printf("File:   %s\n\n", filePath)
		fmt.Printf("✓ File read successfully (%d bytes)\n\n", len(data))
		fmt.Println("=== File Content ===")
		fmt.Println(string(data))
	}
}

func cmdLs() {
	flagSet := flag.NewFlagSet("ls", flag.ExitOnError)
	flagSet.BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	flagSet.StringVar(&logLevel, "log-level", "info", "Log level")
	flagSet.StringVar(&logLevel, "l", "info", "Log level (shorthand)")
	flagSet.Parse(os.Args[2:])

	// 设置日志级别
	if err := logger.SetLevelFromString(logLevel); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if flagSet.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Usage: btrfs-read ls [--json] [-l level] <image> [path]")
		os.Exit(1)
	}

	devicePath := flagSet.Arg(0)
	dirPath := "/"
	if flagSet.NArg() > 1 {
		dirPath = flagSet.Arg(1)
	}

	// 打开文件系统
	filesystem, err := fs.Open(devicePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening filesystem: %v\n", err)
		os.Exit(1)
	}
	defer filesystem.Close()

	// 列出目录
	entries, err := filesystem.ListDirectory(dirPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing directory: %v\n", err)
		os.Exit(1)
	}

	if jsonOutput {
		// JSON 输出格式
		output := map[string]interface{}{
			"path":    dirPath,
			"entries": entries,
		}
		jsonData, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(jsonData))
	} else {
		// 普通文本输出
		fmt.Printf("=== Directory Listing ===\n")
		fmt.Printf("Path: %s\n\n", dirPath)

		if len(entries) == 0 {
			fmt.Println("(empty directory)")
		} else {
			fmt.Printf("%-10s %-15s %s\n", "Type", "Inode", "Name")
			fmt.Println("-------------------------------------------")
			for _, entry := range entries {
				typeStr := getFileTypeName(entry.Type)
				fmt.Printf("%-10s %-15d %s\n", typeStr, entry.Inode, entry.Name)
			}
		}
	}
}

func getFileTypeName(fileType uint8) string {
	switch fileType {
	case 1:
		return "file"
	case 2:
		return "dir"
	case 3:
		return "chrdev"
	case 4:
		return "blkdev"
	case 5:
		return "fifo"
	case 6:
		return "sock"
	case 7:
		return "symlink"
	default:
		return "unknown"
	}
}

func printSuperblockInfo(sb *ondisk.Superblock) {
	fmt.Println("=== Superblock Information ===")
	fmt.Println()

	// 魔数
	fmt.Printf("Magic:           %s", string(sb.Magic[:]))
	if sb.IsValid() {
		fmt.Println(" ✓")
	} else {
		fmt.Println(" ✗ INVALID")
	}

	// 基本信息
	fmt.Printf("Label:           %s\n", sb.GetLabel())
	fmt.Printf("FSID:            %s\n", formatUUID(sb.FSID[:]))
	fmt.Printf("Metadata UUID:   %s\n", formatUUID(sb.MetadataUUID[:]))

	// 大小信息
	fmt.Printf("Total Bytes:     %d (%.2f MB)\n", sb.TotalBytes, float64(sb.TotalBytes)/1024/1024)
	fmt.Printf("Bytes Used:      %d (%.2f MB)\n", sb.BytesUsed, float64(sb.BytesUsed)/1024/1024)
	fmt.Printf("Usage:           %.2f%%\n", float64(sb.BytesUsed)/float64(sb.TotalBytes)*100)

	// 节点大小
	fmt.Printf("Sector Size:     %d bytes\n", sb.SectorSize)
	fmt.Printf("Node Size:       %d bytes\n", sb.NodeSize)
	fmt.Printf("Leaf Size:       %d bytes\n", sb.LeafSize)
	fmt.Printf("Stripe Size:     %d bytes\n", sb.StripeSize)

	// 生成号
	fmt.Printf("Generation:      %d\n", sb.Generation)
	fmt.Printf("Chunk Root Gen:  %d\n", sb.ChunkRootGeneration)

	// 树根地址
	fmt.Printf("\n--- Tree Roots ---\n")
	fmt.Printf("Root Tree:       0x%x (level %d)\n", sb.Root, sb.RootLevel)
	fmt.Printf("Chunk Tree:      0x%x (level %d)\n", sb.ChunkRoot, sb.ChunkRootLevel)
	if sb.LogRoot != 0 {
		fmt.Printf("Log Tree:        0x%x (level %d)\n", sb.LogRoot, sb.LogRootLevel)
	}

	// 设备信息
	fmt.Printf("\n--- Device Information ---\n")
	fmt.Printf("Num Devices:     %d\n", sb.NumDevices)
	fmt.Printf("Device ID:       %d\n", sb.DevItem.DevID)
	fmt.Printf("Device UUID:     %s\n", formatUUID(sb.DevItem.UUID[:]))
	fmt.Printf("Device Total:    %d bytes\n", sb.DevItem.TotalBytes)
	fmt.Printf("Device Used:     %d bytes\n", sb.DevItem.BytesUsed)

	// 校验和类型
	fmt.Printf("\n--- Features ---\n")
	fmt.Printf("Checksum Type:   %s\n", getChecksumTypeName(sb.CsumType))
	fmt.Printf("Compat Flags:    0x%x\n", sb.CompatFlags)
	fmt.Printf("Incompat Flags:  0x%x\n", sb.IncompatFlags)

	// 系统 Chunk 数组
	fmt.Printf("\n--- System Chunk Array ---\n")
	fmt.Printf("Array Size:      %d bytes\n", sb.SysChunkArraySize)
	if sb.SysChunkArraySize > 0 {
		fmt.Printf("Array Data:      %s...\n", hex.EncodeToString(sb.SysChunkArray[:min(32, int(sb.SysChunkArraySize))]))
	}

	fmt.Println()
}

func formatUUID(uuid []byte) string {
	if len(uuid) != 16 {
		return "invalid"
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16])
}

func getChecksumTypeName(csumType uint16) string {
	switch csumType {
	case ondisk.CsumTypeCRC32C:
		return "CRC32C"
	case ondisk.CsumTypeXXHash:
		return "XXHASH"
	case ondisk.CsumTypeSHA256:
		return "SHA256"
	case ondisk.CsumTypeBlake2:
		return "BLAKE2"
	default:
		return fmt.Sprintf("Unknown (%d)", csumType)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
