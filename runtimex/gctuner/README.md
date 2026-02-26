# GCTuner - GC 调优器

动态调整 Go GC 参数以优化内存使用和 GC 性能。

## 功能

- **自动调优**: 根据堆内存使用情况自动调整 GC 频率
- **阈值控制**: 设置堆内存高水位线
- **动态 GCPercent**: 在最小值和最大值之间动态调整
- **全局单例**: 每个进程只允许一个 GC 调优器

## 工作原理

```
 _______________  => limit: 主机/容器内存硬限制
|               |
|---------------| => threshold: 当 gc_trigger < threshold 时增加 GCPercent
|               |
|---------------| => gc_trigger: heap_live + heap_live * GCPercent / 100
|               |
|---------------|
|   heap_live   |
|_______________|
```

Go runtime 只在达到 `gc_trigger` 时触发 GC，通过动态调整 `GCPercent` 可以 tuning GC 性能。

## 使用示例

### 基本用法

```go
package main

import (
    "github.com/yyxxgame/gopkg/runtimex/gctuner"
)

func main() {
    // 设置堆内存阈值（单位：字节）
    // 当堆内存接近此值时，会自动增加 GC 频率
    threshold := uint64(1 << 30) // 1GB
    gctuner.Tuning(threshold)
    
    // 获取当前 GCPercent
    percent := gctuner.GetGCPercent()
    println("当前 GCPercent:", percent)
}
```

### 配置最大/最小 GCPercent

```go
package main

import (
    "fmt"
    "github.com/yyxxgame/gopkg/runtimex/gctuner"
)

func main() {
    // 设置最大 GCPercent（默认 500）
    oldMax := gctuner.SetMaxGCPercent(800)
    fmt.Printf("原最大 GCPercent: %d\n", oldMax)
    
    // 设置最小 GCPercent（默认 50）
    oldMin := gctuner.SetMinGCPercent(20)
    fmt.Printf("原最小 GCPercent: %d\n", oldMin)
    
    // 获取当前配置
    maxPercent := gctuner.GetMaxGCPercent()
    minPercent := gctuner.GetMinGCPercent()
    fmt.Printf("GCPercent 范围：%d - %d\n", minPercent, maxPercent)
}
```

### 禁用调优

```go
// 设置阈值为 0 可禁用 GC 调优
gctuner.Tuning(0)
```

## API 参考

### 函数

#### Tuning

设置 GC 调优器的堆内存阈值。

```go
func Tuning(threshold uint64)
```

**参数**:
- `threshold`: 堆内存阈值（字节）
  - `threshold == 0`: 禁用调优
  - `threshold > 0`: 启用调优

**注意**: 调用此函数后，环境变量 `GOGC` 将不再生效。

#### GetGCPercent

获取当前 GCPercent。

```go
func GetGCPercent() uint32
```

#### GetMaxGCPercent

获取最大 GCPercent 值。

```go
func GetMaxGCPercent() uint32
```

#### SetMaxGCPercent

设置新的最大 GCPercent 值。

```go
func SetMaxGCPercent(n uint32) uint32
```

**返回值**: 原最大 GCPercent 值

#### GetMinGCPercent

获取最小 GCPercent 值。

```go
func GetMinGCPercent() uint32
```

#### SetMinGCPercent

设置新的最小 GCPercent 值。

```go
func SetMinGCPercent(n uint32) uint32
```

**返回值**: 原最小 GCPercent 值

## 测试

```bash
go test -v ./runtimex/gctuner
```

## 配置说明

### 默认值

| 配置项 | 默认值 | 说明 |
|-------|-------|------|
| `minGCPercent` | 50 | 最小 GC 百分比 |
| `maxGCPercent` | 500 | 最大 GC 百分比 |
| `defaultGCPercent` | 100 | 默认 GC 百分比（或 GOGC 环境变量值） |

### GCPercent 含义

`GOGC=100` 表示堆内存增长 100% 时触发 GC。

- **GOGC=50**: 堆增长 50% 触发 GC（更频繁）
- **GOGC=200**: 堆增长 200% 触发 GC（较不频繁）

### 调优策略

1. **内存紧张时**: 降低 GCPercent，增加 GC 频率
2. **CPU 紧张时**: 提高 GCPercent，减少 GC 频率
3. **自动平衡**: 调优器根据堆使用情况自动平衡

## 应用场景

- **内存受限环境**: 容器、K8s Pod
- **大内存应用**: 防止 GC 过于频繁
- **性能敏感应用**: 优化 GC 停顿时间
- **长期运行服务**: 稳定内存使用模式

## 与 GOGC 环境变量对比

| 特性 | GOGC 环境变量 | GCTuner |
|------|------------|---------|
| 动态调整 | ❌ | ✅ |
| 自动适应 | ❌ | ✅ |
| 阈值控制 | ❌ | ✅ |
| 配置复杂度 | 低 | 中 |

## 注意事项

1. **单例限制**: 每个进程只能有一个调优器实例
2. **覆盖 GOGC**: 启用后 `GOGC` 环境变量不再生效
3. **阈值设置**: 应根据实际内存限制设置
4. **监控建议**: 配合内存监控使用效果更佳

## 推荐配置

### 容器环境

```go
// 设置为容器内存限制的 70-80%
containerLimit := 2 * 1024 * 1024 * 1024 // 2GB
gctuner.Tuning(uint64(float64(containerLimit) * 0.75))
```

### 大内存应用

```go
// 设置较高的阈值，减少 GC 频率
gctuner.Tuning(4 * 1024 * 1024 * 1024) // 4GB
gctuner.SetMaxGCPercent(1000) // 允许更高的 GCPercent
```
