package btree

import (
	"encoding/binary"
	"fmt"
)

const (
	HeaderSize = 101
)

// Key B-Tree Key
type Key struct {
	ObjectID uint64
	Type     uint8
	Offset   uint64
}

// Compare compares two keys.
// Returns: -1 (this < other), 0 (==), 1 (this > other).
func (k *Key) Compare(other *Key) int {
	if k.ObjectID < other.ObjectID {
		return -1
	} else if k.ObjectID > other.ObjectID {
		return 1
	}

	if k.Type < other.Type {
		return -1
	} else if k.Type > other.Type {
		return 1
	}

	if k.Offset < other.Offset {
		return -1
	} else if k.Offset > other.Offset {
		return 1
	}

	return 0
}

// Header represents a B-Tree node header.
type Header struct {
	Checksum   [32]byte
	FSID       [16]byte
	Bytenr     uint64
	Flags      uint64
	ChunkUUID  [16]byte
	Generation uint64
	Owner      uint64
	NrItems    uint32
	Level      uint8
}

// IsLeaf reports whether this is a leaf node.
func (h *Header) IsLeaf() bool {
	return h.Level == 0
}

// Node represents a B-Tree node.
type Node struct {
	Header *Header
	Keys   []*Key   // Internal node: keys.
	Ptrs   []uint64 // Internal node: child pointers.
	Items  []*Item  // Leaf node: items.
}

// Item represents an item in a leaf node.
type Item struct {
	Key    *Key
	Offset uint32 // Data offset within the node.
	Size   uint32 // Data size.
	Data   []byte // Actual data.
}

// UnmarshalHeader parses a node header.
func UnmarshalHeader(data []byte) (*Header, error) {
	if len(data) < HeaderSize {
		return nil, fmt.Errorf("header data too short: got %d, need %d", len(data), HeaderSize)
	}

	h := &Header{}
	copy(h.Checksum[:], data[0:32])
	copy(h.FSID[:], data[32:48])
	h.Bytenr = binary.LittleEndian.Uint64(data[48:56])
	h.Flags = binary.LittleEndian.Uint64(data[56:64])
	copy(h.ChunkUUID[:], data[64:80])
	h.Generation = binary.LittleEndian.Uint64(data[80:88])
	h.Owner = binary.LittleEndian.Uint64(data[88:96])
	h.NrItems = binary.LittleEndian.Uint32(data[96:100])
	h.Level = data[100]

	return h, nil
}

// UnmarshalNode parses a node.
func UnmarshalNode(data []byte, nodeSize uint32) (*Node, error) {
	if len(data) < int(nodeSize) {
		return nil, fmt.Errorf("node data too short: got %d, need %d", len(data), nodeSize)
	}

	// Parse header.
	header, err := UnmarshalHeader(data[:HeaderSize])
	if err != nil {
		return nil, err
	}

	node := &Node{Header: header}

	if header.IsLeaf() {
		return unmarshalLeafNode(node, data, nodeSize)
	}
	return unmarshalInternalNode(node, data, nodeSize)
}

func unmarshalLeafNode(node *Node, data []byte, nodeSize uint32) (*Node, error) {
	node.Items = make([]*Item, node.Header.NrItems)

	offset := HeaderSize
	for i := uint32(0); i < node.Header.NrItems; i++ {
		if offset+25 > len(data) { // key(17) + offset(4) + size(4)
			return nil, fmt.Errorf("insufficient data for item %d", i)
		}

		item := &Item{}

		// Parse key.
		key := &Key{}
		key.ObjectID = binary.LittleEndian.Uint64(data[offset:])
		key.Type = data[offset+8]
		key.Offset = binary.LittleEndian.Uint64(data[offset+9:])
		item.Key = key
		offset += 17

		// Parse offset and size.
		item.Offset = binary.LittleEndian.Uint32(data[offset:])
		item.Size = binary.LittleEndian.Uint32(data[offset+4:])
		offset += 8

		node.Items[i] = item
	}

	// Read actual data.
	for _, item := range node.Items {
		dataOffset := HeaderSize + int(item.Offset)
		dataEnd := dataOffset + int(item.Size)

		if dataEnd > len(data) {
			return nil, fmt.Errorf("item data out of bounds: offset=%d, size=%d, nodeSize=%d",
				dataOffset, item.Size, nodeSize)
		}

		item.Data = data[dataOffset:dataEnd]
	}

	return node, nil
}

func unmarshalInternalNode(node *Node, data []byte, nodeSize uint32) (*Node, error) {
	node.Keys = make([]*Key, node.Header.NrItems)
	node.Ptrs = make([]uint64, node.Header.NrItems)

	offset := HeaderSize
	for i := uint32(0); i < node.Header.NrItems; i++ {
		if offset+33 > len(data) { // key(17) + blockptr(8) + generation(8)
			return nil, fmt.Errorf("insufficient data for key_ptr %d", i)
		}

		// Parse key.
		key := &Key{}
		key.ObjectID = binary.LittleEndian.Uint64(data[offset:])
		key.Type = data[offset+8]
		key.Offset = binary.LittleEndian.Uint64(data[offset+9:])
		node.Keys[i] = key
		offset += 17

		// Parse block pointer.
		node.Ptrs[i] = binary.LittleEndian.Uint64(data[offset:])
		offset += 8

		// Skip generation.
		offset += 8
	}

	return node, nil
}
