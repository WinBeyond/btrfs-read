package ondisk

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestSuperblockUnmarshal(t *testing.T) {
	// Create a minimal valid superblock.
	buf := make([]byte, SuperblockSize)

	// Magic is at offset 64 (Checksum:32 + FSID:16 + Bytenr:8 + Flags:8).
	offset := 32 + 16 + 8 + 8
	copy(buf[offset:offset+8], BtrfsMagic[:])

	// Parse.
	sb := &Superblock{}
	err := sb.Unmarshal(buf)

	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Validate magic.
	if !bytes.Equal(sb.Magic[:], BtrfsMagic[:]) {
		t.Errorf("Magic mismatch: got %v, want %v", sb.Magic, BtrfsMagic)
	}

	// Validate IsValid.
	if !sb.IsValid() {
		t.Error("IsValid() returned false for valid superblock")
	}
}

func TestSuperblockInvalidMagic(t *testing.T) {
	buf := make([]byte, SuperblockSize)

	// Set an invalid magic value (at the correct offset).
	offset := 32 + 16 + 8 + 8
	copy(buf[offset:offset+8], []byte("_INVALID"))

	sb := &Superblock{}
	err := sb.Unmarshal(buf)

	if err == nil {
		t.Error("Expected error for invalid magic, got nil")
	}
}

func TestSuperblockTooShort(t *testing.T) {
	buf := make([]byte, 100) // Too short.

	sb := &Superblock{}
	err := sb.Unmarshal(buf)

	if err == nil {
		t.Error("Expected error for short buffer, got nil")
	}
}

func TestSuperblockGetLabel(t *testing.T) {
	sb := &Superblock{}

	// Test empty label.
	label := sb.GetLabel()
	if label != "" {
		t.Errorf("Expected empty label, got %q", label)
	}

	// Test a valid label.
	copy(sb.Label[:], []byte("TestLabel"))
	label = sb.GetLabel()
	if label != "TestLabel" {
		t.Errorf("Expected 'TestLabel', got %q", label)
	}

	// Test label with null terminator.
	copy(sb.Label[:], []byte("Test\x00Extra"))
	label = sb.GetLabel()
	if label != "Test" {
		t.Errorf("Expected 'Test', got %q", label)
	}
}

func TestSuperblockFields(t *testing.T) {
	buf := make([]byte, SuperblockSize)

	// Build a superblock containing various fields.
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

	// Fill the remaining part.
	for w.Len() < SuperblockSize {
		w.WriteByte(0)
	}

	// Parse.
	sb := &Superblock{}
	if err := sb.Unmarshal(buf); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Validate fields.
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
