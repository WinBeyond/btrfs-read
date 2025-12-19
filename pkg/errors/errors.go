package errors

import (
	"errors"
	"fmt"
)

// 预定义错误
var (
	// 设备相关错误
	ErrDeviceNotFound    = errors.New("device not found")
	ErrDeviceReadFailed  = errors.New("device read failed")
	ErrInvalidDeviceSize = errors.New("invalid device size")

	// Superblock 相关错误
	ErrInvalidMagic       = errors.New("invalid magic number")
	ErrInvalidChecksum    = errors.New("invalid checksum")
	ErrNoValidSuperblock  = errors.New("no valid superblock found")
	ErrSuperblockTooSmall = errors.New("superblock data too small")

	// Chunk 相关错误
	ErrChunkNotFound       = errors.New("chunk mapping not found")
	ErrInvalidChunkMapping = errors.New("invalid chunk mapping")
	ErrUnsupportedRaidType = errors.New("unsupported RAID type")

	// B-Tree 相关错误
	ErrNodeNotFound = errors.New("node not found")
	ErrInvalidNode  = errors.New("invalid node data")
	ErrKeyNotFound  = errors.New("key not found")
	ErrInvalidPath  = errors.New("invalid btree path")
	ErrNodeTooSmall = errors.New("node data too small")

	// 文件系统相关错误
	ErrPathNotFound    = errors.New("path not found")
	ErrNotDirectory    = errors.New("not a directory")
	ErrNotRegularFile  = errors.New("not a regular file")
	ErrInodeNotFound   = errors.New("inode not found")
	ErrInvalidFilePath = errors.New("invalid file path")
	ErrExtentNotFound  = errors.New("extent not found")

	// 压缩相关错误
	ErrUnsupportedCompression = errors.New("unsupported compression type")
	ErrDecompressionFailed    = errors.New("decompression failed")

	// 缓存相关错误
	ErrCacheMiss = errors.New("cache miss")
)

// BtrfsError Btrfs 错误包装
type BtrfsError struct {
	Op  string // 操作名称
	Err error  // 原始错误
}

func (e *BtrfsError) Error() string {
	if e.Op == "" {
		return e.Err.Error()
	}
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

func (e *BtrfsError) Unwrap() error {
	return e.Err
}

// Wrap 包装错误
func Wrap(op string, err error) error {
	if err == nil {
		return nil
	}
	return &BtrfsError{
		Op:  op,
		Err: err,
	}
}

// Is 检查错误类型
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As 转换错误类型
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}
