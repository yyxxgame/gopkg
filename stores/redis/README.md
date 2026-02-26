# Redis - Redis 客户端

基于 redis/go-redis/v9 的 Redis 客户端封装。

## 功能

- **连接池**: 内置连接池管理
- **慢查询监控**: 自动记录慢查询
- **分布式追踪**: OpenTelemetry 集成
- **Option 模式**: 灵活的配置选项
- **连接健康检查**: 启动时自动检查连接

## 使用示例

### 基本用法

```go
package main

import (
    "context"
    "github.com/yyxxgame/gopkg/stores/redis"
)

func main() {
    // 创建 Redis 客户端
    rds := redis.NewRedis(redis.RedisConf{
        Host: "localhost:6379",
        Pass: "",
        DB:   0,
    })
    
    ctx := context.Background()
    
    // 设置
    rds.Set(ctx, "key", "value", 0)
    
    // 获取
    val, _ := rds.Get(ctx, "key").Result()
    println(val)
}
```

### 使用 Option 配置

```go
package main

import (
    "time"
    "github.com/yyxxgame/gopkg/stores/redis"
)

func main() {
    rds := redis.NewRedis(
        redis.RedisConf{
            Host: "localhost:6379",
            Pass: "password",
            DB:   1,
        },
        redis.WithDB(2),                    // 覆盖 DB
        redis.WithIdleConns(16),            // 空闲连接数
        redis.WithMaxRetries(5),            // 最大重试次数
        redis.WithReadTimeout(5*time.Second), // 读超时
        redis.WithSlowThresholdDuration(time.Second), // 慢查询阈值
    )
}
```

### 扫描键

```go
package main

import (
    "context"
    "fmt"
    "github.com/yyxxgame/gopkg/stores/redis"
)

func main() {
    rds := redis.NewRedis(redis.RedisConf{
        Host: "localhost:6379",
    })
    
    ctx := context.Background()
    
    // 扫描匹配模式的键
    keys := rds.ScanKeys(ctx, "user:*", 100)
    for _, key := range keys {
        fmt.Println(key)
    }
}
```

### 使用 Redis 命令

```go
// 所有 redis/go-redis/v9 的命令都可以直接使用
rds.Set(ctx, "key", "value", time.Hour)
val := rds.Get(ctx, "key")
rds.Del(ctx, "key")
rds.HSet(ctx, "hash", "field", "value")
rds.LPush(ctx, "list", "item")
// ... 更多命令
```

## API 参考

### 类型

#### Redis

Redis 客户端结构体（嵌入 `*v9rds.Client`）。

```go
type Redis struct {
    *v9rds.Client
    // 其他字段
}
```

#### Option

配置选项函数类型。

```go
type Option func(rds *Redis)
```

### 函数

#### NewRedis

创建新的 Redis 客户端。

```go
func NewRedis(conf RedisConf, opts ...Option) *Redis
```

**参数**:
- `conf`: Redis 配置
- `opts`: 可选配置项

**注意**: 如果连接失败会 panic。

#### ScanKeys

扫描匹配模式的键。

```go
func (rds *Redis) ScanKeys(ctx context.Context, pattern string, count int64) []string
```

**参数**:
- `pattern`: 键模式（如 `user:*`）
- `count`: 每次扫描的数量

**返回值**:
- 匹配的键列表

### 配置选项

#### WithDB

设置数据库编号。

```go
func WithDB(db int) Option
```

#### WithIdleConns

设置最小空闲连接数。

```go
func WithIdleConns(conns int) Option
```

#### WithMaxRetries

设置最大重试次数。

```go
func WithMaxRetries(retries int) Option
```

#### WithReadTimeout

设置读超时时间。

```go
func WithReadTimeout(timeout time.Duration) Option
```

#### WithPingTimeout

设置 ping 超时时间。

```go
func WithPingTimeout(timeout time.Duration) Option
```

#### WithSlowThresholdDuration

设置慢查询阈值。

```go
func WithSlowThresholdDuration(threshold time.Duration) Option
```

#### WithDisableIdentity

禁用 Redis 7.x+ 的客户端 ID 功能（兼容旧版本）。

```go
func WithDisableIdentity() Option
```

#### WithMaintNotificationsConfig

设置维护通知配置。

```go
func WithMaintNotificationsConfig(conf *maintnotifications.Config) Option
```

## 测试

```bash
go test -v ./stores/redis
```

## 内置 Hook

### 1. 慢查询 Hook

自动记录超过阈值的查询：

```go
// 默认阈值：100ms
redis.NewRedis(conf, redis.WithSlowThresholdDuration(100*time.Millisecond))
```

### 2. 追踪 Hook

自动添加 OpenTelemetry Span：

```go
// 默认已启用，与 go-zero trace 集成
```

## 默认配置

| 配置项 | 默认值 |
|-------|-------|
| `maxRetries` | 3 |
| `idleConns` | 8 |
| `slowThresholdDuration` | 100ms |
| `readTimeout` | 2s |
| `pingTimeout` | 1s |

## 应用场景

### 1. 缓存

```go
// 设置缓存
rds.SetEX(ctx, "user:123", userData, 5*time.Minute)

// 获取缓存
data := rds.Get(ctx, "user:123")
```

### 2. 分布式锁

```go
lock := redis.NewLock("lock:resource")
if lock.Acquire() {
    defer lock.Release()
    // 执行临界区代码
}
```

### 3. 计数器

```go
// 自增
rds.Incr(ctx, "counter")

// 自减
rds.Decr(ctx, "counter")
```

### 4. 消息队列

```go
// 入队
rds.LPush(ctx, "queue", message)

// 出队
msg := rds.RPop(ctx, "queue")
```

### 5. 发布/订阅

```go
// 发布
rds.Publish(ctx, "channel", message)

// 订阅
pubsub := rds.Subscribe(ctx, "channel")
```

## 最佳实践

### 1. 键命名规范

```go
// 使用冒号分隔的命名空间
userKey := fmt.Sprintf("user:%d", userID)
orderKey := fmt.Sprintf("order:%d", orderID)
```

### 2. 设置过期时间

```go
// 避免内存泄漏，始终设置过期时间
rds.SetEX(ctx, "cache:key", value, 10*time.Minute)
```

### 3. 批量操作

```go
// 使用 Pipeline 减少网络往返
pipe := rds.Pipeline()
pipe.Set(ctx, "key1", "value1", 0)
pipe.Set(ctx, "key2", "value2", 0)
pipe.Exec(ctx)
```

### 4. 错误处理

```go
val, err := rds.Get(ctx, "key").Result()
if err == redis.Nil {
    // 键不存在
} else if err != nil {
    // 其他错误
}
```

## 注意事项

1. **连接检查**: 初始化时会 ping Redis，失败会 panic
2. **资源管理**: 不需要手动关闭，连接池自动管理
3. **慢查询**: 生产环境设置合适的慢查询阈值
4. **Redis 版本**: 支持 Redis 7.x+，旧版本需使用 `WithDisableIdentity()`
