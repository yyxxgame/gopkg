# gopkg

通用 Go 工具库

[![Go Version](https://img.shields.io/badge/Go-1.25-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## 简介

gopkg 是一个通用的 Go 语言工具库，提供加密、事件处理、消息队列、数据存储、日志记录、分布式追踪和运行时工具等实用功能。

## 主要功能模块

### 🔐 加密 (cryptor)
- AES 加密/解密（CBC/ECB 模式，PKCS7 填充）
- RSA 公钥/私钥加密签名

### 📡 事件总线 (eventbus)
- 观察者模式实现
- 支持异步和同步事件分发

### 📨 消息队列 (mq)
- Kafka 生产者/消费者（基于 IBM/sarama）
- 任务队列支持
- Kafka Q 消费者

### 💾 数据存储 (stores)
- Redis 客户端封装（支持连接池、慢查询监控）
- Elasticsearch 客户端

### 📝 日志 (logw)
- Kafka 日志写入器
- 与 go-zero 日志系统集成

### 🕐 定时任务 (infrastructure/cron)
- v1/v2 版本支持
- 基于 robfig/cron/v3

### 🔄 同步工具 (syncx)
- Goroutine 池（gopool）
- 并发安全集合（ConcurrentMap）

### 🔍 分布式追踪 (xtrace)
- OpenTelemetry 集成
- Jaeger/Zipkin 导出支持

### 📊 监控 (prompusher)
- Prometheus 指标推送器

### ⚙️ 运行时扩展 (runtimex)
- GC 调优器

### 📂 监听器 (watcher)
- 文件监听
- etcd 监听

## 安装

```bash
# 安装依赖
go get -u github.com/yyxxgame/gopkg

# 清理依赖
go mod tidy
```

## 快速开始

### 事件总线

```go
package main

import (
    "github.com/yyxxgame/gopkg/eventbus"
)

func main() {
    bus := eventbus.NewEventBus()
    
    // 订阅事件
    bus.Watch("user.created", func(event *eventbus.Event) {
        fmt.Println("User created:", event.Data)
    })
    
    // 分发事件
    bus.Dispatch(&eventbus.Event{
        Name: "user.created",
        Data: map[string]interface{}{"id": 123},
    })
}
```

### Redis 客户端

```go
package main

import (
    "github.com/yyxxgame/gopkg/stores/redis"
)

func main() {
    rds := redis.NewRedis(redis.RedisConf{
        Host: "localhost:6379",
        Pass: "",
        DB:   0,
    }, redis.WithDB(1), redis.WithIdleConns(16))
    
    // 使用 Redis 客户端
    rds.Set(ctx, "key", "value", 0)
}
```

### Kafka 生产者

```go
package main

import (
    "github.com/yyxxgame/gopkg/mq/saramakafka"
)

func main() {
    producer := saramakafka.NewProducer([]string{"localhost:9092"})
    
    // 发布消息
    err := producer.Publish("topic-name", "key", "message")
    if err != nil {
        log.Fatal(err)
    }
}
```

### 并发 Map

```go
package main

import (
    "github.com/yyxxgame/gopkg/collection/concurrentmap"
)

func main() {
    cm := concurrentmap.NewConcurrentMap[string, int](32)
    
    cm.Set("key1", 100)
    value, ok := cm.Get("key1")
    
    cm.Range(func(key string, value int) bool {
        fmt.Println(key, value)
        return true
    })
}
```

## 核心依赖

- [go-zero](https://github.com/zeromicro/go-zero) - 微服务框架
- [redis/go-redis](https://github.com/redis/go-redis) - Redis 客户端
- [IBM/sarama](https://github.com/IBM/sarama) - Kafka 客户端
- [OpenTelemetry](https://opentelemetry.io) - 分布式追踪
- [Prometheus](https://prometheus.io) - 监控指标

## 开发与测试

```bash
# 构建所有包
go build ./...

# 运行所有测试
go test ./...

# 运行特定测试
go test -run <TestName> ./...

# 生成覆盖率报告
go test -cover ./...

# 格式化代码
go fmt ./...

# 代码检查
go vet ./...
```

## 项目结构

```
gopkg/
├── algorithm/          # 算法实现（backoff）
├── collection/         # 集合工具（concurrentmap）
├── cryptor/            # 加密（AES, RSA）
├── eventbus/           # 事件总线
├── exception/          # 异常处理
├── infrastructure/     # 基础设施（API, cron, queue）
├── internal/           # 内部工具
├── logw/               # 日志工具
├── mq/                 # 消息队列（Kafka）
├── prompusher/         # Prometheus 推送器
├── runtimex/           # 运行时扩展
├── stores/             # 数据存储（Redis, ES）
├── syncx/              # 同步工具
├── watcher/            # 监听器
└── xtrace/             # 分布式追踪
```

## 代码风格

本项目遵循 go-zero 代码风格约定：

- 文件头注释（`//@File`, `//@Time`, `//@Author`）
- 导入分组（标准库 → 本地包 → 第三方包）
- 接口命名前缀 `I`
- Option/Builder 模式
- 工厂函数返回接口

详细规范请参阅 [AGENTS.md](AGENTS.md)。

## License

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！

