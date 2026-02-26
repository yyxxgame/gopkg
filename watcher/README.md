# Watcher - 监听器

基于 etcd 实现的文件/配置监听器。

## 功能

- **etcd 监听**: 监听 etcd 中的键值变化
- **事件通知**: 键值变化时触发回调
- **自动重连**: 连接断开后自动重连
- **分布式同步**: 支持多实例配置同步

## 使用示例

### 基本用法

```go
package main

import (
    "context"
    "fmt"
    "log"
    "github.com/yyxxgame/gopkg/watcher"
)

func main() {
    // etcd 端点
    endpoints := []string{"localhost:2379"}
    
    // 创建监听器
    w, err := watcher.NewWatcher(endpoints, "/config/myapp")
    if err != nil {
        log.Fatal(err)
    }
    defer w.Close()
    
    // 启动监听
    ctx := context.Background()
    err = w.Watch(ctx, func(key, value string) {
        fmt.Printf("配置更新：%s = %s\n", key, value)
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // 阻塞等待事件
    select {}
}
```

### 监听多个键

```go
package main

import (
    "context"
    "fmt"
    "github.com/yyxxgame/gopkg/watcher"
)

func main() {
    endpoints := []string{"localhost:2379"}
    
    // 监听前缀
    w, _ := watcher.NewWatcher(endpoints, "/config/myapp/")
    defer w.Close()
    
    // 监听所有子键的变化
    w.Watch(context.Background(), func(key, value string) {
        // key: /config/myapp/database
        // value: {"host": "localhost", "port": 5432}
        fmt.Printf("配置变化：%s -> %s\n", key, value)
    })
}
```

## API 参考

### 类型

#### Watcher

监听器结构体。

```go
type Watcher struct {
    // 内部字段
}
```

### 函数

#### NewWatcher

创建新的监听器。

```go
func NewWatcher(endpoints []string, key string) (*Watcher, error)
```

**参数**:
- `endpoints`: etcd 集群端点列表
- `key`: 要监听的键或前缀

**返回值**:
- `*Watcher`: 监听器实例
- `error`: 创建错误

#### Watch

启动监听。

```go
func (w *Watcher) Watch(ctx context.Context, handler func(key, value string)) error
```

**参数**:
- `ctx`: 上下文
- `handler`: 事件处理函数

#### Close

关闭监听器。

```go
func (w *Watcher) Close()
```

## 测试

```bash
go test -v ./watcher
```

## 实现细节

- **基于 etcd v3**: 使用 etcd client v3 API
- **Watch API**: 使用 etcd 的 Watch 功能实现实时监听
- **事件类型**: 支持 PUT、DELETE 等操作
- **连接管理**: 自动处理连接断开和重连

## 应用场景

### 1. 配置中心

```go
// 监听配置变化
watcher.NewWatcher(etcdEndpoints, "/config/myapp")

// 配置更新时自动重新加载
```

### 2. 服务发现

```go
// 监听服务列表变化
watcher.NewWatcher(etcdEndpoints, "/services/api")

// 服务上下线时更新路由表
```

### 3. 分布式锁

```go
// 监听锁键变化
watcher.NewWatcher(etcdEndpoints, "/locks/resource1")

// 锁释放时触发业务逻辑
```

### 4. 功能开关

```go
// 监听功能开关状态
watcher.NewWatcher(etcdEndpoints, "/features/new-feature")

// 动态开启/关闭功能
```

## 与文件监听对比

| 特性 | etcd Watcher | 文件监听 |
|------|-------------|---------|
| 分布式支持 | ✅ | ❌ |
| 实时性 | 高 | 中 |
| 可靠性 | 高（etcd 保证） | 中 |
| 配置复杂度 | 中 | 低 |

## 注意事项

1. **etcd 依赖**: 需要部署 etcd 集群
2. **网络连接**: 确保与 etcd 的网络连通性
3. **权限控制**: 配置合适的 etcd 访问权限
4. **资源清理**: 使用完后调用 `Close()` 释放资源

## 最佳实践

### 1. 错误处理

```go
err := w.Watch(ctx, func(key, value string) {
    defer func() {
        if r := recover(); r != nil {
            // 处理回调中的 panic
        }
    }()
    
    // 处理配置更新
})
if err != nil {
    // 处理监听错误
}
```

### 2. 优雅关闭

```go
w, _ := watcher.NewWatcher(endpoints, key)
defer w.Close()

// 监听上下文
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

w.Watch(ctx, handler)
```

### 3. 配置缓存

```go
var configCache atomic.Value

w.Watch(ctx, func(key, value string) {
    // 解析新配置
    newConfig := parseConfig(value)
    // 更新缓存
    configCache.Store(newConfig)
})

// 使用配置
config := configCache.Load().(*Config)
```
