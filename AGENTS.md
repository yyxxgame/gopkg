# AGENTS.md - gopkg 开发指南

## 项目概述

**模块**: `github.com/yyxxgame/gopkg`  
**Go 版本**: 1.25  
**用途**: 通用工具库，提供加密、事件处理、消息队列、数据存储、日志记录、分布式追踪和运行时工具。

**核心依赖**: go-zero, redis/go-redis, IBM/sarama, OpenTelemetry, Prometheus

---

## 开发命令

### 构建
```bash
go build ./...                              # 构建所有包
go build -o <output> <package>              # 构建特定包
go mod tidy                                 # 清理依赖
go mod verify                               # 验证依赖
```

### 测试
```bash
go test ./...                               # 运行所有测试
go test -v ./...                            # 详细输出运行测试
go test -run <TestName> ./...               # 按名称运行特定测试
go test -run <TestName> <package>           # 在特定包中运行单个测试
go test -cover ./...                        # 运行测试并生成覆盖率报告
go test -bench=. ./...                      # 运行基准测试
```

**测试文件**: 共 24 个测试文件，分布在 eventbus、stores/redis、cryptor、mq 等包中

### 代码检查与格式化
```bash
go vet ./...                                # 运行 Go 内置检查器
go fmt ./...                                # 格式化所有 Go 文件
golangci-lint run ./...                     # 运行 golangci-lint（如已安装）
```

**注意**: 项目没有 `.golangci.yml` 配置文件，代码风格遵循 go-zero 约定。

---

## 代码风格指南

### 导入组织

**顺序**: 标准库 → 空行 → 外部包（本地包优先，然后第三方包）

```go
import (
    "context"
    "errors"
    "fmt"
    "sync"
    "time"

    "github.com/yyxxgame/gopkg/syncx/gopool"  // 本地包优先
    "github.com/redis/go-redis/v9"             // 第三方包
    "github.com/zeromicro/go-zero/core/trace"
)
```

### 命名约定

**函数**:
- 导出函数：PascalCase - `NewEventBus()`, `MustNewProducer()`, `GetOrSet()`
- 未导出函数：camelCase - `getShard()`, `calcGCPercent()`, `initResponder()`
- `Must` 前缀用于初始化失败时 panic 的构造函数：`MustNewKafkaWriter()`

**类型**:
- 接口：前缀 `I` - `IBus`, `IProducer`, `IService`, `ITask`
- 结构体：PascalCase，无前缀 - `Redis`, `ConcurrentMap`, `producer`
- 泛型：`type ConcurrentMap[K comparable, V any]`

**变量/常量**:
- camelCase: `defaultPool`, `shardCount`, `maxRetries`
- 短作用域使用短名称：`p` (producer), `ctx` (context), `c` (config)
- 常量：`maxRetries = 3`, `idleConns = 8`

### 错误处理

**标准模式**:
```go
value, err := operation()
if err != nil {
    return nil, err  // 初始化失败时使用 panic(err)
}
```

**自定义错误**:
```go
func MakeError(errCode int32, errStr string) error {
    return errors.New(fmt.Sprintf("%d@%s", errCode, errStr))
}
```

**Panic/Recover** (用于异常处理):
```go
func (sel *TryStruct) Finally(finally func()) {
    defer func() {
        if e := recover(); e != nil {
            // 处理异常
            logx.ErrorStack(e)
        }
        finally()
    }()
    sel.try()
}
```

### 类型定义

**结构体嵌入**:
```go
type producer struct {
    *OptionConf
    hooks []ProducerHook
    sarama.SyncProducer  // 嵌入外部类型
}
```

**接口分组**:
```go
type (
    Callback func(*Event)
    IBus interface {
        Watch(event string, callback Callback)
        Dispatch(event *Event)
    }
    bus struct {
        callbacks sync.Map
    }
)
```

### 注释与文档

**文件头** (每个文件顶部必需):
```go
//@File     filename.go
//@Time     2024/03/19
//@Author   #Suyghur,
```

**函数文档**:
```go
// NewEventBus 创建一个 Bus。
func NewEventBus() IBus { ... }

// GetOrSet 如果键存在则返回其值，否则设置并返回给定值。
func (cm *ConcurrentMap[K, V]) GetOrSet(key K, value V) (actual V, ok bool) { ... }
```

**行内注释**: 实现细节可以使用中文注释：
```go
// 适配 redis 7.x 以下版本
disableIdentity bool
```

### 代码模式

**Option/Builder 模式**:
```go
func NewRedis(conf RedisConf, opts ...Option) *Redis {
    rds := &Redis{ /* 默认值 */ }
    for _, opt := range opts {
        opt(rds)
    }
    return rds
}

func WithDB(db int) Option {
    return func(rds *Redis) { rds.db = db }
}
```

**工厂函数返回接口**:
```go
func NewEventBus() IBus { return &bus{} }
func NewProducer(brokers []string, opts ...Option) IProducer { ... }
```

**全局池模式**:
```go
var defaultPool Pool

func init() {
    defaultPool = NewPool("gopool.DefaultPool", 10000, NewConfig())
}
```

---

## 项目结构

```
gopkg/
├── algorithm/          # 算法实现（backoff）
├── collection/         # 集合工具（concurrentmap）
├── cryptor/            # 加密（AES, RSA）
├── eventbus/           # 事件总线（观察者模式）
├── exception/          # 异常处理工具
├── infrastructure/     # 基础设施（API, cron v1/v2, queue v1/v2）
├── internal/           # 内部工具
├── logw/               # 日志工具（Kafka writer）
├── mq/                 # 消息队列（Kafka: saramakafka, kqkafka, tasks）
├── prompusher/         # Prometheus 指标推送器
├── runtimex/           # 运行时扩展（GC tuner）
├── stores/             # 数据存储（Redis, Elasticsearch）
├── syncx/              # 同步工具（goroutine pool）
├── watcher/            # 文件/etcd watcher
└── xtrace/             # 分布式追踪（OpenTelemetry）
```

---

## Agent 注意事项

1. **无 CI/CD**: 项目没有 GitHub Actions 或 Makefile，使用标准 Go 工具。
2. **文件头**: 始终包含三行文件头注释。
3. **go-zero 约定**: 项目大量使用 go-zero，遵循其模式。
4. **禁止类型压制**: 永远不要使用 `as any` 或 `@ts-ignore` 等价物。
5. **测试覆盖**: 现有 24 个测试文件，确保新代码有测试。
6. **中文注释**: 实现细节可以使用中文，但导出的 API 文档使用英文。
