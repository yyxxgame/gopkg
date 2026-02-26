# Algorithm - 算法库

常用算法实现。

## 功能

### Backoff - 退避算法

实现指数退避（Exponential Backoff）策略，用于重试机制。

**特性**:
- 可配置的基础延迟时间
- 可配置的最大延迟时间
- 可配置的乘数因子
- 支持抖动（Jitter）因子

## 使用示例

```go
package main

import (
    "fmt"
    "time"
    "github.com/yyxxgame/gopkg/algorithm"
)

func main() {
    // 创建退避策略
    backoff := algorithm.NewBackOff(
        algorithm.WithBaseDelay(time.Second),      // 基础延迟 1 秒
        algorithm.WithMaxDelay(120*time.Second),   // 最大延迟 120 秒
        algorithm.WithFactor(2.0),                 // 乘数因子 2.0
        algorithm.WithJitter(1.5),                 // 抖动因子 1.5
    )
    
    // 模拟重试场景
    for i := 0; i < 5; i++ {
        delay := backoff.Duration()
        fmt.Printf("第 %d 次重试，等待 %v\n", i+1, delay)
        time.Sleep(delay)
    }
    
    // 重置退避计数器
    backoff.Reset()
    
    // 获取当前重试次数
    attempts := backoff.Attempts()
    fmt.Printf("当前重试次数：%d\n", attempts)
}
```

## API 参考

### 类型

#### Backoff

退避策略结构体。

```go
type Backoff struct {
    // 内部字段，不直接访问
}
```

#### BackoffOption

配置选项函数类型。

```go
type BackoffOption func(*Backoff)
```

### 函数

#### NewBackOff

创建新的退避策略实例。

```go
func NewBackOff(opts ...BackoffOption) *Backoff
```

**默认配置**:
- `factor`: 2.0
- `jitter`: 1
- `baseDelay`: 1 秒
- `maxDelay`: 120 秒

#### Duration

计算并返回下一次重试的延迟时间。

```go
func (b *Backoff) Duration() time.Duration
```

**计算公式**: `jitter * baseDelay * factor^N(attempts)`

#### Reset

重置退避计数器。

```go
func (b *Backoff) Reset()
```

#### Attempts

获取当前重试次数。

```go
func (b *Backoff) Attempts() int64
```

### 配置选项

#### WithBaseDelay

设置基础延迟时间。

```go
func WithBaseDelay(baseDelay time.Duration) BackoffOption
```

#### WithMaxDelay

设置最大延迟时间。

```go
func WithMaxDelay(maxDelay time.Duration) BackoffOption
```

#### WithFactor

设置乘数因子。

```go
func WithFactor(factor float64) BackoffOption
```

#### WithJitter

设置抖动因子。

```go
func WithJitter(jitter float64) BackoffOption
```

## 测试

```bash
go test -v ./algorithm
```

## 应用场景

- 网络请求重试
- 数据库连接重试
- API 限流后的等待
- 任何需要指数退避策略的场景
