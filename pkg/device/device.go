package device

import (
	"io"
	"os"

	"github.com/WinBeyond/btrfs-read/pkg/errors"
)

// BlockDevice is the block device interface.
type BlockDevice interface {
	// ReadAt reads data at the specified offset.
	ReadAt(p []byte, off int64) (n int, err error)

	// Size returns the device size.
	Size() int64

	// DeviceID returns the unique device identifier.
	DeviceID() uint64

	// Close closes the device.
	Close() error
}

// Ensure FileDevice implements io.ReaderAt.
var _ io.ReaderAt = (*FileDevice)(nil)

// FileDevice is a file-backed block device.
type FileDevice struct {
	file     *os.File
	size     int64
	deviceID uint64
	path     string
}

// NewFileDevice creates a file device.
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
		deviceID: 0, // Will be read from the superblock.
		path:     path,
	}, nil
}

// ReadAt implements io.ReaderAt.
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

// Size returns the device size.
func (d *FileDevice) Size() int64 {
	return d.size
}

// DeviceID returns the device ID.
func (d *FileDevice) DeviceID() uint64 {
	return d.deviceID
}

// SetDeviceID sets the device ID (after reading from the superblock).
func (d *FileDevice) SetDeviceID(id uint64) {
	d.deviceID = id
}

// Path returns the device path.
func (d *FileDevice) Path() string {
	return d.path
}

// Close closes the device.
func (d *FileDevice) Close() error {
	if d.file != nil {
		err := d.file.Close()
		d.file = nil // Set to nil to prevent double close.
		return err
	}
	return nil
}
