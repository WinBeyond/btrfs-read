package btree

import (
	"fmt"
	"sort"
)

// Path B-Tree 搜索路径
type Path struct {
	Nodes []*Node
	Slots []int
}

// NodeReader 节点读取接口
type NodeReader interface {
	ReadNode(logical uint64, nodeSize uint32) (*Node, error)
}

// Searcher B-Tree 搜索器
type Searcher struct {
	reader   NodeReader
	nodeSize uint32
}

// NewSearcher 创建搜索器
func NewSearcher(reader NodeReader, nodeSize uint32) *Searcher {
	return &Searcher{
		reader:   reader,
		nodeSize: nodeSize,
	}
}

// Search 搜索指定的 key
func (s *Searcher) Search(rootAddr uint64, targetKey *Key) (*Path, error) {
	path := &Path{
		Nodes: make([]*Node, 0),
		Slots: make([]int, 0),
	}

	currentAddr := rootAddr

	for {
		// 读取当前节点
		node, err := s.reader.ReadNode(currentAddr, s.nodeSize)
		if err != nil {
			return nil, fmt.Errorf("failed to read node at 0x%x: %w", currentAddr, err)
		}

		// 二分查找
		slot, exact := s.binarySearch(node, targetKey)

		// 对于内部节点，如果没有精确匹配，使用前一个 slot
		// 参考 btrfs-fuse: if (level && ret && slot > 0) slot--;
		if !node.Header.IsLeaf() && !exact && slot > 0 {
			slot--
		}

		path.Nodes = append(path.Nodes, node)
		path.Slots = append(path.Slots, slot)

		// 如果是叶节点，搜索完成
		if node.Header.IsLeaf() {
			return path, nil
		}

		// 内部节点，继续向下
		if slot >= len(node.Ptrs) {
			return nil, fmt.Errorf("invalid slot %d (max %d)", slot, len(node.Ptrs)-1)
		}
		currentAddr = node.Ptrs[slot]
	}
}

// binarySearch 二分查找，返回 (slot, exact_match)
func (s *Searcher) binarySearch(node *Node, targetKey *Key) (int, bool) {
	if node.Header.IsLeaf() {
		return s.binarySearchLeaf(node, targetKey)
	}
	return s.binarySearchInternal(node, targetKey)
}

func (s *Searcher) binarySearchLeaf(node *Node, targetKey *Key) (int, bool) {
	// 找到第一个 key >= targetKey 的位置
	idx := sort.Search(len(node.Items), func(i int) bool {
		return node.Items[i].Key.Compare(targetKey) >= 0
	})

	// 检查是否精确匹配
	exact := false
	if idx < len(node.Items) && node.Items[idx].Key.Compare(targetKey) == 0 {
		exact = true
	}

	return idx, exact
}

func (s *Searcher) binarySearchInternal(node *Node, targetKey *Key) (int, bool) {
	// 找到第一个 key > targetKey 的位置
	idx := sort.Search(len(node.Keys), func(i int) bool {
		return node.Keys[i].Compare(targetKey) > 0
	})

	// 检查前一个是否精确匹配
	exact := false
	if idx > 0 && node.Keys[idx-1].Compare(targetKey) == 0 {
		exact = true
	}

	return idx, exact
}

// GetItem 从 path 获取 item
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

// GetKey 从 path 获取 key（可能精确匹配或下一个）
func (p *Path) GetKey() (*Key, error) {
	item, err := p.GetItem()
	if err != nil {
		return nil, err
	}
	return item.Key, nil
}
