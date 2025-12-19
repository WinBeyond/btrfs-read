package ondisk

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestSuperblockUnmarshal(t *testing.T) {
	// 创建一个最小的有效 superblock
	buf := make([]byte, SuperblockSize)

	// 魔数在偏移 64 (Checksum:32 + FSID:16 + Bytenr:8 + Flags:8)
	offset := 32 + 16 + 8 + 8
	copy(buf[offset:offset+8], BtrfsMagic[:])

	// 解析
	sb := &Superblock{}
	err := sb.Unmarshal(buf)

	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// 验证魔数
	if !bytes.Equal(sb.Magic[:], BtrfsMagic[:]) {
		t.Errorf("Magic mismatch: got %v, want %v", sb.Magic, BtrfsMagic)
	}

	// 验证 IsValid
	if !sb.IsValid() {
		t.Error("IsValid() returned false for valid superblock")
	}
}

func TestSuperblockInvalidMagic(t *testing.T) {
	buf := make([]byte, SuperblockSize)

	// 设置错误的魔数（在正确的偏移位置）
	offset := 32 + 16 + 8 + 8
	copy(buf[offset:offset+8], []byte("_INVALID"))

	sb := &Superblock{}
	err := sb.Unmarshal(buf)

	if err == nil {
		t.Error("Expected error for invalid magic, got nil")
	}
}

func TestSuperblockTooShort(t *testing.T) {
	buf := make([]byte, 100) // 太短

	sb := &Superblock{}
	err := sb.Unmarshal(buf)

	if err == nil {
		t.Error("Expected error for short buffer, got nil")
	}
}

func TestSuperblockGetLabel(t *testing.T) {
	sb := &Superblock{}

	// 测试空标签
	label := sb.GetLabel()
	if label != "" {
		t.Errorf("Expected empty label, got %q", label)
	}

	// 测试有效标签
	copy(sb.Label[:], []byte("TestLabel"))
	label = sb.GetLabel()
	if label != "TestLabel" {
		t.Errorf("Expected 'TestLabel', got %q", label)
	}

	// 测试带 null 终止符的标签
	copy(sb.Label[:], []byte("Test\x00Extra"))
	label = sb.GetLabel()
	if label != "Test" {
		t.Errorf("Expected 'Test', got %q", label)
	}
}

func TestSuperblockFields(t *testing.T) {
	buf := make([]byte, SuperblockSize)

	// 构造一个包含各种字段的 superblock
	w := bytes.NewBuffer(buf[:0])

	// Checksum (32 bytes)
	w.Write(make([]byte, 32))

	// FSID (16 bytes)
	fsid := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	w.Write(fsid)

	// Bytenr
	binary.Write(w, binary.LittleEndian, uint64(0x10000))

	// Flags
	binary.Write(w, binary.LittleEndian, uint64(0))

	// Magic
	w.Write(BtrfsMagic[:])

	// Generation
	binary.Write(w, binary.LittleEndian, uint64(100))

	// Root
	binary.Write(w, binary.LittleEndian, uint64(0x1000000))

	// ChunkRoot
	binary.Write(w, binary.LittleEndian, uint64(0x2000000))

	// LogRoot
	binary.Write(w, binary.LittleEndian, uint64(0))

	// LogRootTransid
	binary.Write(w, binary.LittleEndian, uint64(0))

	// TotalBytes
	binary.Write(w, binary.LittleEndian, uint64(100*1024*1024)) // 100MB

	// BytesUsed
	binary.Write(w, binary.LittleEndian, uint64(10*1024*1024)) // 10MB

	// RootDirObjectid
	binary.Write(w, binary.LittleEndian, uint64(256))

	// NumDevices
	binary.Write(w, binary.LittleEndian, uint64(1))

	// SectorSize
	binary.Write(w, binary.LittleEndian, uint32(4096))

	// NodeSize
	binary.Write(w, binary.LittleEndian, uint32(16384))

	// LeafSize
	binary.Write(w, binary.LittleEndian, uint32(16384))

	// StripeSize
	binary.Write(w, binary.LittleEndian, uint32(4096))

	// 填充剩余部分
	for w.Len() < SuperblockSize {
		w.WriteByte(0)
	}

	// 解析
	sb := &Superblock{}
	if err := sb.Unmarshal(buf); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// 验证字段
	tests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"Bytenr", sb.Bytenr, uint64(0x10000)},
		{"Generation", sb.Generation, uint64(100)},
		{"Root", sb.Root, uint64(0x1000000)},
		{"ChunkRoot", sb.ChunkRoot, uint64(0x2000000)},
		{"TotalBytes", sb.TotalBytes, uint64(100 * 1024 * 1024)},
		{"BytesUsed", sb.BytesUsed, uint64(10 * 1024 * 1024)},
		{"NumDevices", sb.NumDevices, uint64(1)},
		{"SectorSize", sb.SectorSize, uint32(4096)},
		{"NodeSize", sb.NodeSize, uint32(16384)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s: got %v, want %v", tt.name, tt.got, tt.expected)
			}
		})
	}
}

func BenchmarkSuperblockUnmarshal(b *testing.B) {
	buf := make([]byte, SuperblockSize)
	copy(buf[48:56], BtrfsMagic[:])

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sb := &Superblock{}
		sb.Unmarshal(buf)
	}
}
