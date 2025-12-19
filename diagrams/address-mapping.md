# 逻辑地址到物理地址映射流程

```mermaid
flowchart TD
    Start([开始]) --> Input[接收逻辑地址<br/>输入: logical_addr]
    
    Input --> ChunkSearch[在红黑树中查找 Chunk<br/>搜索条件:<br/>chunk.start ≤ logical_addr<br/>&& logical_addr < chunk.start + chunk.length]
    
    ChunkSearch --> ChunkFound{找到 Chunk?}
    ChunkFound -->|否| Error1[返回错误<br/>ErrChunkNotFound]
    Error1 --> Stop1([结束])
    
    ChunkFound -->|是| CalcOffset[计算 Chunk 内偏移<br/>offset_in_chunk =<br/>logical_addr - chunk.logical_start]
    
    CalcOffset --> RaidType{RAID 类型?}
    
    RaidType -->|SINGLE/DUP| Single[直接映射<br/>physical =<br/>chunk.stripes0.offset<br/>+ offset_in_chunk]
    
    RaidType -->|RAID0| Raid0[条带化计算<br/>stripe_nr = offset_in_chunk / stripe_len<br/>stripe_idx = stripe_nr % num_stripes<br/>stripe_offset = stripe_nr / num_stripes * stripe_len<br/>offset_in_stripe = offset_in_chunk % stripe_len<br/><br/>physical = chunk.stripesstripe_idx.offset<br/>+ stripe_offset + offset_in_stripe]
    
    RaidType -->|RAID1| Raid1[镜像选择<br/>返回所有镜像地址:<br/>for each stripe:<br/>physicali = stripe.offset + offset_in_chunk<br/><br/>通常选择第一个镜像]
    
    RaidType -->|RAID10| Raid10[RAID1+0 组合<br/>先 RAID0 条带化<br/>再 RAID1 镜像]
    
    RaidType -->|RAID5/6| Raid56[奇偶校验计算<br/>计算数据条带和奇偶条带<br/>处理条带旋转<br/>复杂实现]
    
    RaidType -->|未知| Error2[返回错误<br/>ErrUnsupportedRaidType]
    Error2 --> Stop2([结束])
    
    Single --> Output
    Raid0 --> Output
    Raid1 --> Output
    Raid10 --> Output
    Raid56 --> Output
    
    Output[返回物理地址<br/>PhysicalAddr:<br/>- device_id: uint64<br/>- offset: uint64]
    Output --> Stop3([结束])
```

## RAID 类型说明

| RAID 类型 | 特点 |
|-----------|------|
| SINGLE | 单设备，无冗余 |
| DUP | 同设备双份 (元数据) |
| RAID0 | 条带化，提高性能 |
| RAID1 | 镜像，提高可靠性 |
| RAID10 | RAID1+0 组合 |
| RAID5 | 单奇偶校验 |
| RAID6 | 双奇偶校验 |
