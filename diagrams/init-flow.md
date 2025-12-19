# Btrfs Filesystem Initialization

```mermaid
sequenceDiagram
    actor User
    participant FS as FileSystem
    participant VM as VolumeManager
    participant SR as SuperblockReader
    participant BD as BlockDevice
    participant CM as ChunkManager
    participant BTS as BTreeSearcher

    User->>FS: Open(devicePaths)
    activate FS
    
    Note over FS,BTS: 1. 设备扫描
    FS->>VM: ScanDevices(paths)
    activate VM
    loop 遍历每个设备路径
        VM->>BD: NewFileDevice(path)
        activate BD
        BD-->>VM: device
        deactivate BD
        VM->>VM: 存储设备 (deviceID -> device)
    end
    VM-->>FS: volumeManager
    deactivate VM
    
    Note over FS,BTS: 2. 读取 Superblock
    FS->>SR: NewSuperblockReader(device)
    activate SR
    FS->>SR: ReadLatest()
    
    Note right of SR: 读取主 Superblock (64KB)
    SR->>BD: ReadAt(0x10000, 4096)
    activate BD
    BD-->>SR: data[4096]
    deactivate BD
    
    SR->>SR: Unmarshal(data)
    SR->>SR: VerifyMagic("_BHRfS_M")
    SR->>SR: VerifyChecksum(CRC32C)
    
    alt Superblock 有效
        SR-->>FS: superblock
    else 主 Superblock 损坏
        Note right of SR: 尝试备份 Superblock (64MB)
        SR->>BD: ReadAt(0x4000000, 4096)
        BD-->>SR: backup_data
        SR->>SR: 验证备份
        SR-->>FS: superblock (backup)
    end
    deactivate SR
    
    Note over FS,BTS: 3. 初始化 Chunk Tree
    FS->>CM: NewChunkManager()
    activate CM
    CM-->>FS: chunkManager
    deactivate CM
    
    FS->>CM: LoadFromTree(superblock.ChunkRoot)
    activate CM
    
    CM->>BTS: Search(chunkRoot, CHUNK_ITEM key)
    activate BTS
    BTS->>BD: ReadNode(chunkRoot)
    BD-->>BTS: node_data
    BTS->>BTS: ParseNode()
    BTS->>BTS: BinarySearch(CHUNK_ITEM)
    BTS-->>CM: path (叶节点 + slot)
    deactivate BTS
    
    loop 遍历所有 CHUNK_ITEM
        Note right of CM: 解析:<br/>- logical_start<br/>- length<br/>- RAID type<br/>- stripe info
        CM->>CM: ParseChunkItem(item.Data)
        
        Note right of CM: 插入红黑树
        CM->>CM: AddMapping(chunkMapping)
        
        CM->>BTS: NextItem()
        BTS-->>CM: next_item
    end
    
    CM-->>FS: 完成 Chunk Tree 加载
    deactivate CM
    
    Note over FS,BTS: 4. 定位 FS Tree
    FS->>BTS: Search(rootTree, FS_TREE_OBJECTID)
    activate BTS
    BTS->>CM: LogicalToPhysical(rootTree)
    CM-->>BTS: physicalAddr
    BTS->>BD: ReadAt(physicalAddr)
    BD-->>BTS: node_data
    BTS->>BTS: ParseNode()
    BTS->>BTS: FindKey(FS_TREE_OBJECTID)
    BTS-->>FS: fsTreeRoot (逻辑地址)
    deactivate BTS
    
    Note over FS,BTS: 5. 初始化缓存
    Note right of FS: LRU 缓存，256 个块
    FS->>FS: NewBlockCache(256)
    
    Note over FS,BTS: 完成
    FS-->>User: fileSystem
    deactivate FS
    
    Note over User,BD: 初始化完成后，文件系统已准备好:<br/>- 所有设备已加载<br/>- Superblock 已解析<br/>- Chunk 映射已建立<br/>- FS Tree 已定位<br/>- 缓存已初始化
```

## 初始化步骤说明

### 1. 设备扫描
- 遍历所有设备路径
- 为每个设备创建 BlockDevice 实例
- 建立 deviceID 到 device 的映射

### 2. 读取 Superblock
- 主 Superblock 位于偏移 0x10000 (64KB)
- 验证魔数 "_BHRfS_M"
- 验证 CRC32C 校验和
- 如果主 Superblock 损坏,尝试读取备份 (偏移 0x4000000, 64MB)

### 3. 初始化 Chunk Tree
- 从 Superblock 获取 Chunk Tree 根节点地址
- 遍历 Chunk Tree 中的所有 CHUNK_ITEM
- 解析每个 Chunk 的映射信息:
  - logical_start: 逻辑起始地址
  - length: Chunk 长度
  - RAID type: SINGLE/DUP/RAID0/1/5/6/10
  - stripe info: 条带信息
- 将映射插入红黑树以便快速查找

### 4. 定位 FS Tree
- 在 Root Tree 中搜索 FS_TREE_OBJECTID
- 获取 FS Tree 的根节点逻辑地址
- FS Tree 包含所有文件和目录的元数据

### 5. 初始化缓存
- 创建 LRU 缓存 (256 个块)
- 减少重复的磁盘 I/O 操作
- 提高文件读取性能

## 初始化后状态

初始化完成后,文件系统处于就绪状态:
- ✅ 所有设备已加载
- ✅ Superblock 已解析
- ✅ Chunk 映射已建立 (逻辑→物理地址转换)
- ✅ FS Tree 已定位
- ✅ LRU 缓存已初始化
