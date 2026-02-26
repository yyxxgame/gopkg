# gopkg

通用 Go 工具库

[![Go Version](https://img.shields.io/badge/Go-1.25-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## 简介

gopkg 是一个通用的 Go 语言工具库，提供加密、事件处理、消息队列、数据存储、日志记录、分布式追踪和运行时工具等实用功能。

## 主要功能模块

### 🔐 加密 (cryptor)

- **AES 加密** (`cryptor/aes/`) - CBC/ECB 模式，PKCS7 填充，GCM 认证加密 [查看文档 →](cryptor/aes/README.md)
- **RSA 加密** (`cryptor/rsa/`) - 公钥/私钥加密签名 [查看文档 →](cryptor/rsa/README.md)

### 📡 事件总线 (eventbus)

观察者模式实现，支持异步/同步事件分发 [查看文档 →](eventbus/README.md)

### 📨 消息队列 (mq)

- **SaramaKafka** (`mq/saramakafka/`) - Kafka 生产者/消费者，基于 IBM/sarama [查看文档 →](mq/saramakafka/README.md)
- **KqKafka** (`mq/kqkafka/`) - 基于 go-queue 的消费者 [查看文档 →](mq/kqkafka/README.md)
- **Tasks** (`mq/tasks/`) - 任务队列核心 [查看文档 →](mq/tasks/README.md)

### 💾 数据存储 (stores)

- **Redis** (`stores/redis/`) - 基于 redis/go-redis/v9，连接池管理，慢查询监控 [查看文档 →](stores/redis/README.md)
- **Elasticsearch** (`stores/elastic/`) - 基于 olivere/elastic/v7 [查看文档 →](stores/elastic/README.md)

### 📝 日志 (logw)

Kafka 日志写入器，与 go-zero 日志系统集成 [查看文档 →](logw/README.md)

### 🕐 定时任务 (infrastructure/cron)

- **Cron v2** (`infrastructure/cron/v2/`) - 基于 robfig/cron/v3，支持标准 cron 表达式 [查看文档 →](infrastructure/cron/v2/README.md)

### 🔄 任务队列 (infrastructure/queue)

- **Queue v1** (`infrastructure/queue/`) - 基础任务队列 [查看文档 →](infrastructure/queue/README.md)
- **Queue v2** (`infrastructure/queue/v2/`) - 增强版任务队列 [查看文档 →](infrastructure/queue/v2/README.md)

### 🔄 同步工具 (syncx)

- **GoPool** (`syncx/gopool/`) - Goroutine 池，Panic 恢复，Context 支持 [查看文档 →](syncx/gopool/README.md)
- **ConcurrentMap** (`collection/concurrentmap/`) - 并发安全 Map，分片锁机制 [查看文档 →](collection/concurrentmap/README.md)

### 🔍 分布式追踪 (xtrace)

基于 OpenTelemetry，支持 Jaeger、Zipkin、OTLP 导出 [查看文档 →](xtrace/README.md)

### 📊 监控 (prompusher)

Prometheus 指标推送器

### ⚙️ 运行时扩展 (runtimex)

- **GCTuner** (`runtimex/gctuner/`) - 动态调整 GC 参数，堆内存阈值控制 [查看文档 →](runtimex/gctuner/README.md)

### 📂 监听器 (watcher)

基于 etcd 的监听器，配置变更通知 [查看文档 →](watcher/README.md)

### 🧮 算法 (algorithm)

- **Backoff** (`algorithm/`) - 指数退避算法 [查看文档 →](algorithm/README.md)

### 🛡️ 异常处理 (exception)

Panic/Recover 工具

## 安装

```bash
go get -u github.com/yyxxgame/gopkg
go mod tidy
```

## 项目结构

```
gopkg/
├── algorithm/              # 算法实现（backoff）
├── collection/             # 集合工具（concurrentmap）
├── cryptor/                # 加密（AES, RSA）
├── eventbus/               # 事件总线
├── exception/              # 异常处理
├── infrastructure/         # 基础设施（API, cron, queue）
├── internal/               # 内部工具
├── logw/                   # 日志工具
├── mq/                     # 消息队列（Kafka）
├── prompusher/             # Prometheus 推送器
├── runtimex/               # 运行时扩展
├── stores/                 # 数据存储（Redis, ES）
├── syncx/                  # 同步工具
├── watcher/                # 监听器
└── xtrace/                 # 分布式追踪
```

## 文档导航

### 按功能查找

| 功能 | 包路径 | 文档 |
|------|--------|------|
| 加密解密 | `cryptor/aes/` | [AES 文档](cryptor/aes/README.md) |
| 加密解密 | `cryptor/rsa/` | [RSA 文档](cryptor/rsa/README.md) |
| 事件驱动 | `eventbus/` | [事件总线文档](eventbus/README.md) |
| 消息队列 | `mq/saramakafka/` | [Kafka 文档](mq/saramakafka/README.md) |
| 数据存储 | `stores/redis/` | [Redis 文档](stores/redis/README.md) |
| 数据存储 | `stores/elastic/` | [ES 文档](stores/elastic/README.md) |
| 定时任务 | `infrastructure/cron/v2/` | [Cron 文档](infrastructure/cron/v2/README.md) |
| 任务队列 | `infrastructure/queue/v2/` | [Queue 文档](infrastructure/queue/v2/README.md) |
| 并发编程 | `syncx/gopool/` | [GoPool 文档](syncx/gopool/README.md) |
| 并发编程 | `collection/concurrentmap/` | [ConcurrentMap 文档](collection/concurrentmap/README.md) |
| 分布式追踪 | `xtrace/` | [追踪文档](xtrace/README.md) |
| 日志记录 | `logw/` | [日志文档](logw/README.md) |
| GC 调优 | `runtimex/gctuner/` | [GC 调优文档](runtimex/gctuner/README.md) |
| 配置监听 | `watcher/` | [监听器文档](watcher/README.md) |
| 退避算法 | `algorithm/` | [退避算法文档](algorithm/README.md) |

## 代码风格

本项目遵循 go-zero 代码风格约定，详细规范请参阅 **[AGENTS.md](AGENTS.md)**。

## License

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！
