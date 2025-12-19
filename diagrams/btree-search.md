# B-Tree Search Algorithm

```mermaid
flowchart TD
    Start([开始]) --> Input[接收搜索请求<br/>输入:<br/>- root_addr: 根节点逻辑地址<br/>- target_key: 目标 Key]
    
    Input --> InitPath[初始化 Path<br/>path = nodes: [], slots: []]
    
    InitPath --> SetCurrent[current_addr = root_addr]
    
    SetCurrent --> ReadNode[读取当前节点<br/>1. 逻辑地址 → 物理地址<br/>2. 从设备读取数据<br/>3. 验证校验和]
    
    ReadNode --> ParseHeader[解析节点头<br/>header: checksum, fsid,<br/>bytenr, level, nritems]
    
    ParseHeader --> CheckLevel{level == 0?}
    
    CheckLevel -->|是 - 叶节点| ParseItems[解析 items<br/>每个 item:<br/>- key: BtrfsKey<br/>- offset: 数据偏移<br/>- size: 数据大小<br/>- data: 实际数据]
    
    ParseItems --> LeafBinarySearch[二分查找 target_key<br/>left = 0, right = nritems]
    
    LeafBinarySearch --> LeafLoop{left < right?}
    LeafLoop -->|是| LeafMid[mid = left + right / 2<br/>cmp = itemsmid.key.Compare]
    LeafMid --> LeafCmp{cmp < 0?}
    LeafCmp -->|是| LeafLeft[left = mid + 1]
    LeafCmp -->|否| LeafRight[right = mid]
    LeafLeft --> LeafLoop
    LeafRight --> LeafLoop
    
    LeafLoop -->|否| LeafSlot[slot = left]
    LeafSlot --> LeafRecord[记录到 Path<br/>path.nodes.append<br/>path.slots.append]
    LeafRecord --> Return[返回 Path]
    Return --> Stop([结束])
    
    CheckLevel -->|否 - 内部节点| ParseKeyPtrs[解析 key_ptrs<br/>每个 key_ptr:<br/>- key: BtrfsKey<br/>- blockptr: 子节点地址<br/>- generation: 生成号]
    
    ParseKeyPtrs --> InternalBinarySearch[二分查找 target_key<br/>left = 0, right = nritems]
    
    InternalBinarySearch --> InternalLoop{left < right?}
    InternalLoop -->|是| InternalMid[mid = left + right / 2<br/>cmp = key_ptrsmid.key.Compare]
    InternalMid --> InternalCmp{cmp < 0?}
    InternalCmp -->|是| InternalLeft[left = mid + 1]
    InternalCmp -->|否| InternalRight[right = mid]
    InternalLeft --> InternalLoop
    InternalRight --> InternalLoop
    
    InternalLoop -->|否| InternalSlot[slot = left]
    InternalSlot --> AdjustSlot{slot > 0 &&<br/>没有精确匹配?}
    AdjustSlot -->|是| DecSlot[slot = slot - 1<br/>使用前一个指针]
    AdjustSlot -->|否| InternalRecord
    DecSlot --> InternalRecord[记录到 Path<br/>path.nodes.append<br/>path.slots.append]
    
    InternalRecord --> NextNode[current_addr =<br/>key_ptrsslot.blockptr<br/>递归到子节点]
    NextNode --> ReadNode
```

## Key 比较规则

```
func (k *Key) Compare(other *Key) int {
    if k.objectid != other.objectid {
        return k.objectid - other.objectid
    }
    if k.type != other.type {
        return k.type - other.type
    }
    return k.offset - other.offset
}
```

比较顺序:
1. 比较 objectid
2. 如果相等，比较 type
3. 如果相等，比较 offset

## 节点类型说明

| 节点类型 | level | 内容 |
|----------|-------|------|
| 内部节点 | > 0 | key_ptrs[] (指向子节点) |
| 叶节点 | = 0 | items[] (实际数据) |

## Path 结构

- `nodes`: 从 root 到 leaf 的所有节点
- `slots`: 每个节点中的 slot 索引

例如查找 key=(256, INODE_ITEM, 0):
```
path = {
    nodes: [root_node, internal_node, leaf_node]
    slots: [3, 5, 12]
}
```
