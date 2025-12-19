package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/WinBeyond/btrfs-read/pkg/ondisk"
)

const (
	testImagePath    = "../testdata/test.img"
	superblockOffset = 0x10000
)

func TestReadSuperblockFromImage(t *testing.T) {
	// Check whether the test image exists.
	imagePath, err := filepath.Abs(testImagePath)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		t.Skip("Test image not found. Run 'make create-test-image' first.")
	}

	t.Logf("Reading test image: %s", imagePath)

	// Open image file.
	file, err := os.Open(imagePath)
	if err != nil {
		t.Fatalf("Failed to open image: %v", err)
	}
	defer file.Close()

	// Read superblock.
	buf := make([]byte, ondisk.SuperblockSize)
	n, err := file.ReadAt(buf, superblockOffset)
	if err != nil {
		t.Fatalf("Failed to read superblock: %v", err)
	}
	if n != ondisk.SuperblockSize {
		t.Fatalf("Read %d bytes, expected %d", n, ondisk.SuperblockSize)
	}

	// Parse superblock.
	sb := &ondisk.Superblock{}
	if err := sb.Unmarshal(buf); err != nil {
		t.Fatalf("Failed to unmarshal superblock: %v", err)
	}

	// Validate magic.
	if !sb.IsValid() {
		t.Errorf("Invalid magic number: %v", sb.Magic)
	}

	// Validate basic fields.
	t.Run("BasicFields", func(t *testing.T) {
		if sb.SectorSize == 0 {
			t.Error("SectorSize is 0")
		}
		if sb.NodeSize == 0 {
			t.Error("NodeSize is 0")
		}
		if sb.TotalBytes == 0 {
			t.Error("TotalBytes is 0")
		}
	})

	// Validate label.
	t.Run("Label", func(t *testing.T) {
		label := sb.GetLabel()
		t.Logf("Label: %s", label)
		// The test script sets the label to "TestBtrfs".
		if label != "TestBtrfs" {
			t.Logf("Warning: Expected label 'TestBtrfs', got '%s'", label)
		}
	})

	// Validate tree root addresses.
	t.Run("TreeRoots", func(t *testing.T) {
		if sb.Root == 0 {
			t.Error("Root tree address is 0")
		}
		if sb.ChunkRoot == 0 {
			t.Error("Chunk root address is 0")
		}
		t.Logf("Root tree: 0x%x", sb.Root)
		t.Logf("Chunk root: 0x%x", sb.ChunkRoot)
	})

	// Validate device info.
	t.Run("DeviceInfo", func(t *testing.T) {
		if sb.NumDevices == 0 {
			t.Error("NumDevices is 0")
		}
		if sb.DevItem.TotalBytes == 0 {
			t.Error("DevItem.TotalBytes is 0")
		}
		t.Logf("Num devices: %d", sb.NumDevices)
		t.Logf("Device total: %d bytes", sb.DevItem.TotalBytes)
	})

	// Print full info (verbose mode only).
	if testing.Verbose() {
		printSuperblockInfo(t, sb)
	}
}

func TestReadBackupSuperblock(t *testing.T) {
	imagePath, err := filepath.Abs(testImagePath)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		t.Skip("Test image not found. Run 'make create-test-image' first.")
	}

	file, err := os.Open(imagePath)
	if err != nil {
		t.Fatalf("Failed to open image: %v", err)
	}
	defer file.Close()

	// Get file size.
	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}
	fileSize := stat.Size()

	// Backup superblock offsets.
	backupOffsets := []struct {
		name   string
		offset int64
	}{
		{"Primary", 0x10000},
		{"Backup1", 0x4000000},    // 64MB
		{"Backup2", 0x4000000000}, // 256GB (may be beyond file size)
	}

	for _, backup := range backupOffsets {
		t.Run(backup.name, func(t *testing.T) {
			if backup.offset >= fileSize {
				t.Skipf("Offset %d beyond file size %d", backup.offset, fileSize)
			}

			buf := make([]byte, ondisk.SuperblockSize)
			n, err := file.ReadAt(buf, backup.offset)
			if err != nil {
				t.Errorf("Failed to read at offset 0x%x: %v", backup.offset, err)
				return
			}
			if n != ondisk.SuperblockSize {
				t.Errorf("Read %d bytes, expected %d", n, ondisk.SuperblockSize)
				return
			}

			sb := &ondisk.Superblock{}
			if err := sb.Unmarshal(buf); err != nil {
				t.Logf("Unmarshal failed (expected for some backups): %v", err)
				return
			}

			if sb.IsValid() {
				t.Logf("✓ Valid superblock at offset 0x%x, generation %d", backup.offset, sb.Generation)
			}
		})
	}
}

func BenchmarkReadSuperblock(b *testing.B) {
	imagePath, err := filepath.Abs(testImagePath)
	if err != nil {
		b.Fatalf("Failed to get absolute path: %v", err)
	}

	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		b.Skip("Test image not found")
	}

	file, err := os.Open(imagePath)
	if err != nil {
		b.Fatalf("Failed to open image: %v", err)
	}
	defer file.Close()

	buf := make([]byte, ondisk.SuperblockSize)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		file.ReadAt(buf, superblockOffset)
		sb := &ondisk.Superblock{}
		sb.Unmarshal(buf)
	}
}

// Helper: print superblock details.
func printSuperblockInfo(t *testing.T, sb *ondisk.Superblock) {
	t.Logf("\n=== Superblock Information ===")
	t.Logf("Magic:         %s", string(sb.Magic[:]))
	t.Logf("Label:         %s", sb.GetLabel())
	t.Logf("Generation:    %d", sb.Generation)
	t.Logf("Total Bytes:   %d (%.2f MB)", sb.TotalBytes, float64(sb.TotalBytes)/1024/1024)
	t.Logf("Bytes Used:    %d (%.2f MB)", sb.BytesUsed, float64(sb.BytesUsed)/1024/1024)
	t.Logf("Sector Size:   %d", sb.SectorSize)
	t.Logf("Node Size:     %d", sb.NodeSize)
	t.Logf("Num Devices:   %d", sb.NumDevices)
	t.Logf("Root Tree:     0x%x (level %d)", sb.Root, sb.RootLevel)
	t.Logf("Chunk Root:    0x%x (level %d)", sb.ChunkRoot, sb.ChunkRootLevel)
}
