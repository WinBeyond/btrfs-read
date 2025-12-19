package errors

import (
	"errors"
	"fmt"
)

// Predefined errors.
var (
	// Device-related errors.
	ErrDeviceNotFound    = errors.New("device not found")
	ErrDeviceReadFailed  = errors.New("device read failed")
	ErrInvalidDeviceSize = errors.New("invalid device size")

	// Superblock-related errors.
	ErrInvalidMagic       = errors.New("invalid magic number")
	ErrInvalidChecksum    = errors.New("invalid checksum")
	ErrNoValidSuperblock  = errors.New("no valid superblock found")
	ErrSuperblockTooSmall = errors.New("superblock data too small")

	// Chunk-related errors.
	ErrChunkNotFound       = errors.New("chunk mapping not found")
	ErrInvalidChunkMapping = errors.New("invalid chunk mapping")
	ErrUnsupportedRaidType = errors.New("unsupported RAID type")

	// B-Tree-related errors.
	ErrNodeNotFound = errors.New("node not found")
	ErrInvalidNode  = errors.New("invalid node data")
	ErrKeyNotFound  = errors.New("key not found")
	ErrInvalidPath  = errors.New("invalid btree path")
	ErrNodeTooSmall = errors.New("node data too small")

	// Filesystem-related errors.
	ErrPathNotFound    = errors.New("path not found")
	ErrNotDirectory    = errors.New("not a directory")
	ErrNotRegularFile  = errors.New("not a regular file")
	ErrInodeNotFound   = errors.New("inode not found")
	ErrInvalidFilePath = errors.New("invalid file path")
	ErrExtentNotFound  = errors.New("extent not found")

	// Compression-related errors.
	ErrUnsupportedCompression = errors.New("unsupported compression type")
	ErrDecompressionFailed    = errors.New("decompression failed")

	// Cache-related errors.
	ErrCacheMiss = errors.New("cache miss")
)

// BtrfsError wraps a Btrfs error.
type BtrfsError struct {
	Op  string // Operation name.
	Err error  // Original error.
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

// Wrap wraps an error.
func Wrap(op string, err error) error {
	if err == nil {
		return nil
	}
	return &BtrfsError{
		Op:  op,
		Err: err,
	}
}

// Is checks the error type.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As converts the error type.
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}
