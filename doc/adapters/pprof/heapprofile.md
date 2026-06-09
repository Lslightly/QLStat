# Heap Profile

Go Heap Profile 是基于 [pprof protobuf 格式](profile.proto) 的特化实现，核心实现在 Go 标准库 `runtime/pprof/protomem.go` 的 `writeHeapProto` 函数中。与 CPU profile 等通用 profile 类型相比，Go Heap Profile 有以下特化之处：

## 1. 多值类型（Multi-value SampleType）

通用 pprof profile 通常只包含一个 sample type（如 CPU profile 只有 `["cpu", "nanoseconds"]`），但 Go Heap Profile **一个 profile 中包含 4 个 sample type**，每个 sample 对应 4 个 int64 值：

| index | type           | unit    | 含义                                   |
| ----- | -------------- | ------- | -------------------------------------- |
| 0     | `alloc_objects` | `count` | 分配的对象数（采样调整后）             |
| 1     | `alloc_space`   | `bytes` | 分配的空间大小（采样调整后）           |
| 2     | `inuse_objects` | `count` | 当前存留的对象数（采样调整后）         |
| 3     | `inuse_space`   | `bytes` | 当前存留的空间大小（采样调整后）       |

```go
b.pbValueType(tagProfile_SampleType, "alloc_objects", "count")
b.pbValueType(tagProfile_SampleType, "alloc_space", "bytes")
b.pbValueType(tagProfile_SampleType, "inuse_objects", "count")
b.pbValueType(tagProfile_SampleType, "inuse_space", "bytes")
```

这种设计使得一个 profile 同时承载 分配总量 和 当前存活量 两种视角，用户无需分别采集两个 profile。

## 2. 采样周期语义（Period Semantics）

通用 pprof 的 `period_type` / `period` 描述采样事件的单位与间隔。Go Heap Profile 的 period 语义为：

- **PeriodType**: `("space", "bytes")` — 表示采样基于分配空间触发
- **Period**: `rate` — 即 `runtime.MemProfileRate` 的值（默认 512 KB），表示平均每分配这么多字节触发一次采样

这与 CPU profile 的 `("cpu", "nanoseconds")` + 纳秒间隔的语义不同——CPU period 是时间单位，Heap period 是空间单位。

## 3. 采样缩放（Poisson Scaling）

Heap profile 是采样数据，单个 sample 的值不能直接作为真实分配量，必须经过 **Poisson 采样调整**。

函数 `scaleHeapSample(count, size, rate)` 实现了逆概率加权：

```go
avgSize := float64(size) / float64(count)
scale := 1 / (1 - math.Exp(-avgSize/float64(rate)))
return int64(float64(count) * scale), int64(float64(size) * scale)
```

数学原理：
- 对于采样率 `rate=R`，每次分配大小为 `size` 时，被采样到的概率为 `P = 1 - exp(-size/R)`
- 因此，观测到的 `count` 和 `size` 只是真实分配量的 `P` 倍，估计真实值（即全采样时应有的值）为：
  - `真实分配次数 ≈ count / P`
  - `真实分配空间 ≈ size / P`
- 将当前 sample 中所有分配视为同等大小（avgSize = size/count），则每个分配被采样的概率相同，统一缩放即可。反过来，当通过 `go tool pprof -inuse_space` 等工具看 profile 时，实际值是经过逆缩放后的估计值。

**特殊情况**：
- `rate == 1`：全采样，不做调整
- `rate < 1`：视为未知，不做调整

## 4. 运行时栈帧过滤（Runtime Frame Filtering）

Go Heap Profile 默认隐藏 `runtime.` 和 `internal/runtime/` 前缀的运行时内部函数，其逻辑分两次尝试：

1. **第一次（hideRuntime = true）**：扫描 stack，找到第一个非 runtime 函数后，截断之前的 runtime 帧；仅保留非 runtime 帧及其**之上的** runtime 帧（即运行时调用链中暴露给用户代码的部分）。
2. **第二次（hideRuntime = false）**：如果第一次过滤后 `appendLocsForStack` 返回空的 location 列表（说明全是 runtime 帧），则回退到**显示所有帧**，确保 profile 不会空。

```go
hideRuntime := true
for tries := 0; tries < 2; tries++ {
    if hideRuntime {
        for i, addr := range stk {
            if f := runtime.FuncForPC(addr); f != nil &&
                (strings.HasPrefix(f.Name(), "runtime.") || strings.HasPrefix(f.Name(), "internal/runtime/")) {
                continue
            }
            stk = stk[i:]
            break
        }
    }
    locs = b.appendLocsForStack(locs[:0], stk)
    if len(locs) > 0 {
        break
    }
    hideRuntime = false
}
```

这种设计在保持 profile 可读性（不淹没在 runtime 内部函数中）和完整性（不返回空 profile）之间做了平衡。相比之下，CPU profile 通常不做此类过滤或过滤策略不同。

## 5. 逐 Sample 的 BlockSize Label

每个 heap sample 携带一个数值标签 `bytes`（带单位 `bytes`），表示该 allocation 的**平均块大小**（blockSize = AllocBytes / AllocObjects）：

```go
var blockSize int64
if r.AllocObjects > 0 {
    blockSize = r.AllocBytes / r.AllocObjects
}
b.pbSample(values, locs, func() {
    if blockSize != 0 {
        b.pbLabel(tagSample_Label, "bytes", "", blockSize)
    }
})
```

对应 protobuf 的 `Label` 消息：
```protobuf
message Label {
  int64 key = 1;      // "bytes" → string_table index
  int64 num = 3;      // blockSize
  int64 num_unit = 4; // "bytes" → string_table index
}
```

该标签可用于 `go tool pprof` 的 `--functions` 视图中按分配粒度（而非仅总量）排序分析。

## 6. 与 protobuf schema 的映射总结

| Profile Proto 字段      | Go Heap Profile 映射                                                                |
| ----------------------- | ----------------------------------------------------------------------------------- |
| `sample_type[]`         | 4 个 `(type, unit)`：alloc_objects/count, alloc_space/bytes, inuse_objects/count, inuse_space/bytes |
| `sample[].value[]`      | 4 个 int64，对应上述 4 个 sample type                                                |
| `sample[].location_id[]`| 调用栈（经过 runtime 帧过滤）                                                         |
| `sample[].label[]`      | 一个 `bytes` 标签（数值 + 单位），表示平均块大小                                      |
| `period_type`           | `("space", "bytes")`                                                               |
| `period`                | `runtime.MemProfileRate`（采样率）                                                   |
| `default_sample_type`   | 可选，由调用者指定，缺省时 pprof 工具默认选最后一个 sample_type                       |
| `mapping` / `location` / `function` / `string_table` | 标准填充，与其他 profile 类型一致 |

## 7. 与其他 Profile 类型的差异总结

| 维度             | CPU Profile                  | Heap Profile                                  |
| ---------------- | ---------------------------- | --------------------------------------------- |
| Value 数量       | 1（cpu/纳秒）                | 4（alloc/inuse × objects/bytes）              |
| 采样周期类型     | ("cpu", "nanoseconds")       | ("space", "bytes")                            |
| 采样调整         | 无（直接计数值）             | Poisson 逆概率缩放                            |
| 运行时帧过滤     | 通常不做                     | 默认隐藏 runtime 帧，空栈时回退               |
| Sample 标签      | 无                           | 每个 sample 带 `bytes` 块大小标签             |
| 分析视角         | 时间热点                     | 分配热点（总量 / 存活量 / 块大小）            |
