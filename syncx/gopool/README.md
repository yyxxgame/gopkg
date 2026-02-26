# GoPool - Goroutine 池

高性能 goroutine 池实现，支持 panic 恢复和上下文管理。

## 功能

- **Goroutine 复用**: 减少 goroutine 创建销毁开销
- **Panic 恢复**: 自动捕获和处理 panic
- **上下文支持**: 支持 context.Context 管理
- **多池管理**: 支持注册和获取多个池
- **可配置容量**: 可设置池容量

## 使用示例

### 基本用法

```go
package main

import (
    "fmt"
    "github.com/yyxxgame/gopkg/syncx/gopool"
)

func main() {
    // 使用默认池（推荐）
    gopool.Go(func() {
        fmt.Println("任务在 goroutine 池中执行")
        // panic 会被自动捕获
    })
    
    // 获取当前工作协程数
    count := gopool.WorkerCount()
    fmt.Printf("当前工作协程数：%d\n", count)
}
```

### 使用 Context

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/yyxxgame/gopkg/syncx/gopool"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    gopool.CtxGo(ctx, func() {
        // 任务会监听 context 取消信号
        select {
        case <-ctx.Done():
            fmt.Println("任务被取消")
            return
        default:
            fmt.Println("执行任务")
        }
    })
}
```

### 自定义 Panic 处理

```go
package main

import (
    "context"
    "fmt"
    "github.com/yyxxgame/gopkg/syncx/gopool"
)

func main() {
    // 设置全局 panic 处理器
    gopool.SetPanicHandler(func(ctx context.Context, panic interface{}) {
        fmt.Printf("捕获到 panic: %v\n", panic)
        // 可以在这里记录日志或发送告警
    })
    
    gopool.Go(func() {
        panic("测试 panic")
    })
    
    // panic 会被自定义处理器处理
}
```

### 多池管理

```go
package main

import (
    "fmt"
    "github.com/yyxxgame/gopkg/syncx/gopool"
)

func main() {
    // 创建自定义池
    customPool := gopool.NewPool("MyPool", 1000, gopool.NewConfig())
    
    // 注册池
    err := gopool.RegisterPool(customPool)
    if err != nil {
        panic(err)
    }
    
    // 获取已注册的池
    pool := gopool.GetPool("MyPool")
    if pool != nil {
        pool.CtxGo(context.Background(), func() {
            fmt.Println("在自定义池中执行")
        })
    }
}
```

## API 参考

### 全局函数

#### Go

提交任务到默认池。

```go
func Go(f func())
```

**注意**: 推荐使用 `CtxGo` 替代。

#### CtxGo

提交任务到默认池（支持 context）。

```go
func CtxGo(ctx context.Context, f func())
```

**参数**:
- `ctx`: 上下文
- `f`: 任务函数

#### SetCap

设置默认池容量（不推荐）。

```go
func SetCap(cap int32)
```

**注意**: 此函数会影响全局池，不推荐使用。

#### SetPanicHandler

设置全局 panic 处理器。

```go
func SetPanicHandler(f func(context.Context, interface{}))
```

#### WorkerCount

获取当前工作协程数。

```go
func WorkerCount() int32
```

#### RegisterPool

注册新的池到全局映射。

```go
func RegisterPool(p Pool) error
```

**返回值**:
- 如果名称已存在则返回错误

#### GetPool

根据名称获取已注册的池。

```go
func GetPool(name string) Pool
```

**返回值**:
- 如果不存在则返回 nil

### Pool 接口

```go
type Pool interface {
    Name() string
    CtxGo(ctx context.Context, f func())
    SetCap(cap int32)
    SetPanicHandler(f func(context.Context, interface{}))
    WorkerCount() int32
}
```

### 配置

#### Config

池配置。

```go
type Config struct {
    // 可配置字段
}

func NewConfig() *Config
```

#### NewPool

创建新的池。

```go
func NewPool(name string, cap int32, config *Config) Pool
```

## 测试

```bash
go test -v ./syncx/gopool
```

## 性能优势

| 特性 | gopool | 原生 goroutine |
|------|--------|---------------|
| Goroutine 复用 | ✅ | ❌ |
| Panic 恢复 | ✅ | ❌ |
| 容量限制 | ✅ | ❌ |
| 性能 | 高并发下更优 | 低并发下相当 |

## 默认池配置

- **名称**: `gopool.DefaultPool`
- **容量**: 10000
- **Panic 处理**: 自动记录日志

## 应用场景

- **高并发任务**: 大量短生命周期任务
- **Web 请求处理**: HTTP 请求并发控制
- **批量数据处理**: 数据导入导出
- **定时任务**: 后台任务执行
- **任何需要限制并发数的场景**

## 注意事项

1. **不要滥用**: 简单场景使用原生 goroutine 即可
2. **Context 使用**: 优先使用 `CtxGo` 以便控制任务生命周期
3. **Panic 处理**: 设置合适的 panic 处理器记录异常
4. **池容量**: 根据系统资源合理设置容量
