package btree

import (
	"fmt"
	"sort"
)

// Path represents a B-Tree search path.
type Path struct {
	Nodes []*Node
	Slots []int
}

// NodeReader is the node read interface.
type NodeReader interface {
	ReadNode(logical uint64, nodeSize uint32) (*Node, error)
}

// Searcher is a B-Tree searcher.
type Searcher struct {
	reader   NodeReader
	nodeSize uint32
}

// NewSearcher creates a searcher.
func NewSearcher(reader NodeReader, nodeSize uint32) *Searcher {
	return &Searcher{
		reader:   reader,
		nodeSize: nodeSize,
	}
}

// Search searches for the specified key.
func (s *Searcher) Search(rootAddr uint64, targetKey *Key) (*Path, error) {
	path := &Path{
		Nodes: make([]*Node, 0),
		Slots: make([]int, 0),
	}

	currentAddr := rootAddr

	for {
		// Read current node.
		node, err := s.reader.ReadNode(currentAddr, s.nodeSize)
		if err != nil {
			return nil, fmt.Errorf("failed to read node at 0x%x: %w", currentAddr, err)
		}

		// Binary search.
		slot, exact := s.binarySearch(node, targetKey)

		// For internal nodes, if not exact match, use the previous slot.
		// See btrfs-fuse: if (level && ret && slot > 0) slot--;
		if !node.Header.IsLeaf() && !exact && slot > 0 {
			slot--
		}

		path.Nodes = append(path.Nodes, node)
		path.Slots = append(path.Slots, slot)

		// If it's a leaf node, search is complete.
		if node.Header.IsLeaf() {
			return path, nil
		}

		// Internal node, continue downward.
		if slot >= len(node.Ptrs) {
			return nil, fmt.Errorf("invalid slot %d (max %d)", slot, len(node.Ptrs)-1)
		}
		currentAddr = node.Ptrs[slot]
	}
}

// binarySearch performs binary search, returning (slot, exact_match).
func (s *Searcher) binarySearch(node *Node, targetKey *Key) (int, bool) {
	if node.Header.IsLeaf() {
		return s.binarySearchLeaf(node, targetKey)
	}
	return s.binarySearchInternal(node, targetKey)
}

func (s *Searcher) binarySearchLeaf(node *Node, targetKey *Key) (int, bool) {
	// Find the first key >= targetKey.
	idx := sort.Search(len(node.Items), func(i int) bool {
		return node.Items[i].Key.Compare(targetKey) >= 0
	})

	// Check for exact match.
	exact := false
	if idx < len(node.Items) && node.Items[idx].Key.Compare(targetKey) == 0 {
		exact = true
	}

	return idx, exact
}

func (s *Searcher) binarySearchInternal(node *Node, targetKey *Key) (int, bool) {
	// Find the first key > targetKey.
	idx := sort.Search(len(node.Keys), func(i int) bool {
		return node.Keys[i].Compare(targetKey) > 0
	})

	// Check whether the previous key is an exact match.
	exact := false
	if idx > 0 && node.Keys[idx-1].Compare(targetKey) == 0 {
		exact = true
	}

	return idx, exact
}

// GetItem gets an item from the path.
func (p *Path) GetItem() (*Item, error) {
	if len(p.Nodes) == 0 {
		return nil, fmt.Errorf("empty path")
	}

	leafNode := p.Nodes[len(p.Nodes)-1]
	slot := p.Slots[len(p.Slots)-1]

	if !leafNode.Header.IsLeaf() {
		return nil, fmt.Errorf("last node in path is not a leaf")
	}

	if slot >= len(leafNode.Items) {
		return nil, fmt.Errorf("slot %d out of range (max %d)", slot, len(leafNode.Items)-1)
	}

	return leafNode.Items[slot], nil
}

// GetKey gets the key from the path (exact match or next).
func (p *Path) GetKey() (*Key, error) {
	item, err := p.GetItem()
	if err != nil {
		return nil, err
	}
	return item.Key, nil
}
