package device

import (
	"io"
	"os"

	"github.com/yourname/btrfs-read/pkg/errors"
)

// BlockDevice 块设备接口
type BlockDevice interface {
	// ReadAt 在指定偏移处读取数据
	ReadAt(p []byte, off int64) (n int, err error)

	// Size 返回设备大小
	Size() int64

	// DeviceID 返回设备唯一标识
	DeviceID() uint64

	// Close 关闭设备
	Close() error
}

// 确保实现了 io.ReaderAt 接口
var _ io.ReaderAt = (*FileDevice)(nil)

// FileDevice 文件后端块设备
type FileDevice struct {
	file     *os.File
	size     int64
	deviceID uint64
	path     string
}

// NewFileDevice 创建文件设备
func NewFileDevice(path string) (*FileDevice, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap("NewFileDevice", err)
	}

	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, errors.Wrap("NewFileDevice.Stat", err)
	}

	if stat.Size() == 0 {
		file.Close()
		return nil, errors.Wrap("NewFileDevice", errors.ErrInvalidDeviceSize)
	}

	return &FileDevice{
		file:     file,
		size:     stat.Size(),
		deviceID: 0, // 将从 superblock 读取
		path:     path,
	}, nil
}

// ReadAt 实现 io.ReaderAt 接口
func (d *FileDevice) ReadAt(p []byte, off int64) (int, error) {
	if off < 0 {
		return 0, errors.Wrap("FileDevice.ReadAt", io.EOF)
	}
	if off >= d.size {
		return 0, errors.Wrap("FileDevice.ReadAt", io.EOF)
	}

	n, err := d.file.ReadAt(p, off)
	if err != nil && err != io.EOF {
		return n, errors.Wrap("FileDevice.ReadAt", err)
	}

	return n, err
}

// Size 返回设备大小
func (d *FileDevice) Size() int64 {
	return d.size
}

// DeviceID 返回设备 ID
func (d *FileDevice) DeviceID() uint64 {
	return d.deviceID
}

// SetDeviceID 设置设备 ID（从 superblock 读取后设置）
func (d *FileDevice) SetDeviceID(id uint64) {
	d.deviceID = id
}

// Path 返回设备路径
func (d *FileDevice) Path() string {
	return d.path
}

// Close 关闭设备
func (d *FileDevice) Close() error {
	if d.file != nil {
		err := d.file.Close()
		d.file = nil // 设置为 nil 防止重复关闭
		return err
	}
	return nil
}
