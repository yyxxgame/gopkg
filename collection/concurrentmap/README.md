# ConcurrentMap - 并发安全的 Map

基于分片技术实现的线程安全的并发 Map。

## 特性

- **并发安全**: 支持多 goroutine 同时读写
- **分片锁**: 使用 32 个分片（可配置），减少锁竞争
- **泛型支持**: 支持任意可比较的键类型
- **丰富操作**: 支持 Set、Get、Delete、Range 等操作

## 使用示例

```go
package main

import (
    "fmt"
    "github.com/yyxxgame/gopkg/collection/concurrentmap"
)

func main() {
    // 创建并发 Map（默认 32 个分片）
    cm := concurrentmap.NewConcurrentMap[string, int](32)
    
    // 设置值
    cm.Set("key1", 100)
    cm.Set("key2", 200)
    
    // 获取值
    if value, ok := cm.Get("key1"); ok {
        fmt.Printf("key1 = %d\n", value)
    }
    
    // GetOrSet: 如果键存在则返回其值，否则设置并返回给定值
    actual, ok := cm.GetOrSet("key3", 300)
    fmt.Printf("key3 = %d, exists = %v\n", actual, ok)
    
    // 检查键是否存在
    if cm.Has("key1") {
        fmt.Println("key1 exists")
    }
    
    // 删除键
    cm.Delete("key2")
    
    // GetAndDelete: 获取并删除
    if val, ok := cm.GetAndDelete("key1"); ok {
        fmt.Printf("deleted key1 = %d\n", val)
    }
    
    // 遍历所有键值对
    cm.Range(func(key string, value int) bool {
        fmt.Printf("%s: %d\n", key, value)
        return true // 返回 false 可停止遍历
    })
    
    // 获取所有键
    keys := cm.Keys()
    fmt.Printf("keys: %v\n", keys)
    
    // 获取所有值
    values := cm.Values()
    fmt.Printf("values: %v\n", values)
    
    // 获取键数量
    size := cm.KeySize()
    fmt.Printf("size: %d\n", size)
}
```

## API 参考

### 类型

#### ConcurrentMap

```go
type ConcurrentMap[K comparable, V any] struct {
    // 内部字段，不直接访问
}
```

### 函数

#### NewConcurrentMap

创建新的并发 Map。

```go
func NewConcurrentMap[K comparable, V any](shardCount int) *ConcurrentMap[K, V]
```

**参数**:
- `shardCount`: 分片数量，如果 <= 0 则使用默认值 32

#### Set

设置键值对。

```go
func (cm *ConcurrentMap[K, V]) Set(key K, value V)
```

#### Get

获取键对应的值。

```go
func (cm *ConcurrentMap[K, V]) Get(key K) (V, bool)
```

**返回值**:
- `value`: 键对应的值
- `ok`: 键是否存在

#### GetOrSet

获取或设置键值对。

```go
func (cm *ConcurrentMap[K, V]) GetOrSet(key K, value V) (actual V, ok bool)
```

**返回值**:
- `actual`: 实际的值（已存在的或新设置的）
- `ok`: 键是否已存在

#### Delete

删除键值对。

```go
func (cm *ConcurrentMap[K, V]) Delete(key K)
```

#### GetAndDelete

获取并删除键值对。

```go
func (cm *ConcurrentMap[K, V]) GetAndDelete(key K) (actual V, ok bool)
```

#### Has

检查键是否存在。

```go
func (cm *ConcurrentMap[K, V]) Has(key K) bool
```

#### Range

遍历所有键值对。

```go
func (cm *ConcurrentMap[K, V]) Range(iterator func(key K, value V) bool)
```

**参数**:
- `iterator`: 遍历函数，返回 false 时停止遍历

#### Keys

获取所有键。

```go
func (cm *ConcurrentMap[K, V]) Keys() []K
```

#### Values

获取所有值。

```go
func (cm *ConcurrentMap[K, V]) Values() []V
```

#### KeySize

获取键的数量。

```go
func (cm *ConcurrentMap[K, V]) KeySize() int
```

## 测试

```bash
go test -v ./collection/concurrentmap
```

## 性能优化

- **分片锁机制**: 将 Map 分成多个分片，每个分片有独立的锁，减少锁竞争
- **读写锁**: 使用 `sync.RWMutex`，读操作不阻塞
- **默认 32 分片**: 适合大多数场景，可根据实际情况调整

## 与 sync.Map 的对比

| 特性 | ConcurrentMap | sync.Map |
|------|--------------|----------|
| 分片锁 | ✅ 可配置 | ❌ 固定 |
| 泛型支持 | ✅ | ❌ |
| 遍历支持 | ✅ Range | ✅ Range |
| 获取所有键/值 | ✅ | ❌ |
| 获取大小 | ✅ | ❌ |

## 应用场景

- 缓存系统
- 会话存储
- 并发计数器
- 任何需要并发安全的键值存储场景
