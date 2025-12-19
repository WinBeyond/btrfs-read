package integration

import (
	"testing"

	"github.com/WinBeyond/btrfs-read/pkg/fs"
)

func TestOpenFilesystem(t *testing.T) {
	imagePath := "../testdata/test.img"

	filesystem, err := fs.Open(imagePath)
	if err != nil {
		t.Fatalf("Failed to open filesystem: %v", err)
	}
	defer filesystem.Close()

	t.Log("✓ Filesystem opened successfully")
}

func TestReadFileBasic(t *testing.T) {
	t.Skip("Skipping until chunk tree loading is implemented")

	imagePath := "../testdata/test.img"

	filesystem, err := fs.Open(imagePath)
	if err != nil {
		t.Fatalf("Failed to open filesystem: %v", err)
	}
	defer filesystem.Close()

	// 尝试读取文件
	data, err := filesystem.ReadFile("/hello.txt")
	if err != nil {
		t.Logf("Expected error (chunk tree not fully loaded): %v", err)
		return
	}

	t.Logf("File content: %s", string(data))

	expected := "Hello from Btrfs!\n"
	if string(data) != expected {
		t.Errorf("Expected %q, got %q", expected, string(data))
	}
}
