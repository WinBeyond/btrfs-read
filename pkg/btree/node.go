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

// Compare 比较两个 Key
// 返回: -1 (this < other), 0 (==), 1 (this > other)
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

// Header B-Tree 节点头
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

// IsLeaf 是否是叶节点
func (h *Header) IsLeaf() bool {
	return h.Level == 0
}

// Node B-Tree 节点
type Node struct {
	Header *Header
	Keys   []*Key   // 内部节点: keys
	Ptrs   []uint64 // 内部节点: 子节点指针
	Items  []*Item  // 叶节点: items
}

// Item 叶节点中的 item
type Item struct {
	Key    *Key
	Offset uint32 // 数据在节点中的偏移
	Size   uint32 // 数据大小
	Data   []byte // 实际数据
}

// UnmarshalHeader 解析节点头
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

// UnmarshalNode 解析节点
func UnmarshalNode(data []byte, nodeSize uint32) (*Node, error) {
	if len(data) < int(nodeSize) {
		return nil, fmt.Errorf("node data too short: got %d, need %d", len(data), nodeSize)
	}

	// 解析 header
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

		// 解析 key
		key := &Key{}
		key.ObjectID = binary.LittleEndian.Uint64(data[offset:])
		key.Type = data[offset+8]
		key.Offset = binary.LittleEndian.Uint64(data[offset+9:])
		item.Key = key
		offset += 17

		// 解析 offset 和 size
		item.Offset = binary.LittleEndian.Uint32(data[offset:])
		item.Size = binary.LittleEndian.Uint32(data[offset+4:])
		offset += 8

		node.Items[i] = item
	}

	// 读取实际数据
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

		// 解析 key
		key := &Key{}
		key.ObjectID = binary.LittleEndian.Uint64(data[offset:])
		key.Type = data[offset+8]
		key.Offset = binary.LittleEndian.Uint64(data[offset+9:])
		node.Keys[i] = key
		offset += 17

		// 解析 block pointer
		node.Ptrs[i] = binary.LittleEndian.Uint64(data[offset:])
		offset += 8

		// 跳过 generation
		offset += 8
	}

	return node, nil
}
