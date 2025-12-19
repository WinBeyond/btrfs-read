package device

import (
	"os"
	"testing"
)

func TestNewFileDevice(t *testing.T) {
	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "btrfs-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// 写入一些数据
	testData := []byte("test data for btrfs device")
	if _, err := tmpFile.Write(testData); err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	tmpFile.Close()

	// 测试打开设备
	device, err := NewFileDevice(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create device: %v", err)
	}
	defer device.Close()

	// 验证大小
	if device.Size() != int64(len(testData)) {
		t.Errorf("Expected size %d, got %d", len(testData), device.Size())
	}

	// 验证路径
	if device.Path() != tmpFile.Name() {
		t.Errorf("Expected path %s, got %s", tmpFile.Name(), device.Path())
	}
}

func TestFileDevice_ReadAt(t *testing.T) {
	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "btrfs-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// 写入测试数据
	testData := []byte("0123456789ABCDEFGHIJ")
	if _, err := tmpFile.Write(testData); err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	tmpFile.Close()

	// 打开设备
	device, err := NewFileDevice(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create device: %v", err)
	}
	defer device.Close()

	// 测试从头读取
	buf := make([]byte, 5)
	n, err := device.ReadAt(buf, 0)
	if err != nil {
		t.Errorf("ReadAt failed: %v", err)
	}
	if n != 5 {
		t.Errorf("Expected to read 5 bytes, got %d", n)
	}
	if string(buf) != "01234" {
		t.Errorf("Expected '01234', got '%s'", string(buf))
	}

	// 测试从中间读取
	n, err = device.ReadAt(buf, 10)
	if err != nil {
		t.Errorf("ReadAt failed: %v", err)
	}
	if n != 5 {
		t.Errorf("Expected to read 5 bytes, got %d", n)
	}
	if string(buf) != "ABCDE" {
		t.Errorf("Expected 'ABCDE', got '%s'", string(buf))
	}

	// 测试读取超出范围
	_, err = device.ReadAt(buf, int64(len(testData)))
	if err == nil {
		t.Error("Expected error when reading beyond EOF")
	}
}

func TestFileDevice_SetDeviceID(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "btrfs-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Write([]byte("test"))
	tmpFile.Close()

	device, err := NewFileDevice(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create device: %v", err)
	}
	defer device.Close()

	// 初始 ID 应该是 0
	if device.DeviceID() != 0 {
		t.Errorf("Expected initial device ID 0, got %d", device.DeviceID())
	}

	// 设置新 ID
	device.SetDeviceID(12345)
	if device.DeviceID() != 12345 {
		t.Errorf("Expected device ID 12345, got %d", device.DeviceID())
	}
}

func TestNewFileDevice_Errors(t *testing.T) {
	// 测试不存在的文件
	_, err := NewFileDevice("/nonexistent/file/path")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}

	// 测试空文件
	tmpFile, err := os.CreateTemp("", "btrfs-empty-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	_, err = NewFileDevice(tmpFile.Name())
	if err == nil {
		t.Error("Expected error for empty file")
	}
}

func TestFileDevice_Close(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "btrfs-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Write([]byte("test"))
	tmpFile.Close()

	device, err := NewFileDevice(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create device: %v", err)
	}

	// 关闭设备
	if err := device.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// 再次关闭应该不报错
	if err := device.Close(); err != nil {
		t.Errorf("Second close failed: %v", err)
	}
}

func BenchmarkFileDevice_ReadAt(b *testing.B) {
	tmpFile, err := os.CreateTemp("", "btrfs-bench-*")
	if err != nil {
		b.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// 写入 1MB 数据
	data := make([]byte, 1024*1024)
	tmpFile.Write(data)
	tmpFile.Close()

	device, err := NewFileDevice(tmpFile.Name())
	if err != nil {
		b.Fatalf("Failed to create device: %v", err)
	}
	defer device.Close()

	buf := make([]byte, 4096)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		device.ReadAt(buf, int64(i%1000)*4096)
	}
}
