# SaramaKafka - Kafka 生产者/消费者

基于 IBM/sarama 的 Kafka 客户端实现。

## 功能

### 生产者 (Producer)
- **同步发送**: 等待 Kafka 确认
- **追踪集成**: 支持 OpenTelemetry 追踪
- **Hook 机制**: 支持自定义钩子函数
- **分区策略**: 支持轮询等分区策略

### 消费者 (Consumer)
- **消费者组**: 支持消费者组负载均衡
- **自动提交**: 自动提交 offset
- **消息处理**: 支持批量消息处理

## 使用示例

### 生产者

```go
package main

import (
    "github.com/yyxxgame/gopkg/mq/saramakafka"
)

func main() {
    // 创建生产者
    producer := saramakafka.NewProducer(
        []string{"localhost:9092"},
        saramakafka.WithUsername("user"),
        saramakafka.WithPassword("pass"),
    )
    defer producer.Release()
    
    // 发送消息
    err := producer.Publish(
        "topic-name",
        "message-key",
        "message-payload",
    )
    if err != nil {
        panic(err)
    }
}
```

### 使用 Context 的生产者

```go
package main

import (
    "context"
    "github.com/yyxxgame/gopkg/mq/saramakafka"
)

func main() {
    producer := saramakafka.NewProducer([]string{"localhost:9092"})
    defer producer.Release()
    
    ctx := context.Background()
    
    // 使用 Context 发送消息（支持追踪）
    err := producer.PublishCtx(
        ctx,
        "topic-name",
        "key",
        "payload",
    )
    if err != nil {
        panic(err)
    }
}
```

### 消费者

```go
package main

import (
    "context"
    "fmt"
    "github.com/yyxxgame/gopkg/mq/saramakafka"
)

func main() {
    // 创建消费者
    consumer := saramakafka.NewConsumer(
        []string{"localhost:9092"},
        "consumer-group",
        []string{"topic-name"},
    )
    defer consumer.Close()
    
    // 消费消息
    ctx := context.Background()
    err := consumer.Consume(ctx, func(msg *sarama.ConsumerMessage) error {
        fmt.Printf("收到消息：%s = %s\n", string(msg.Key), string(msg.Value))
        return nil
    })
    if err != nil {
        panic(err)
    }
}
```

### 添加自定义 Hook

```go
package main

import (
    "context"
    "fmt"
    "github.com/yyxxgame/gopkg/mq/saramakafka"
)

func main() {
    // 添加日志 hook
    producer := saramakafka.NewProducer(
        []string{"localhost:9092"},
        saramakafka.WithProducerHook(func(ctx context.Context, topic, key, payload string, next saramakafka.ProducerHook) error {
            fmt.Printf("发送消息到 %s: %s\n", topic, payload)
            return next(ctx, topic, key, payload)
        }),
    )
    defer producer.Release()
    
    producer.Publish("topic", "key", "payload")
}
```

## API 参考

### 生产者接口

```go
type IProducer interface {
    Publish(topic, key, payload string) error
    PublishCtx(ctx context.Context, topic, key, payload string) error
    Release()
}
```

### 消费者接口

```go
type IConsumer interface {
    Consume(ctx context.Context, handler func(*sarama.ConsumerMessage) error) error
    Close()
}
```

### 配置选项

#### 生产者选项

```go
// 设置用户名/密码
func WithUsername(username string) Option
func WithPassword(password string) Option

// 设置分区器
func WithPartitioner(partitioner sarama.PartitionerConstructor) Option

// 添加 Hook
func WithProducerHook(hook ProducerHook) Option
func WithProducerHooks(hooks ...ProducerHook) Option
```

#### 消费者选项

```go
// 设置消费者组配置
func WithConsumerGroupConf(conf sarama.Config) Option
```

### Hook 类型

```go
type ProducerHook func(ctx context.Context, topic, key, payload string, next ProducerHook) error
```

## 测试

```bash
go test -v ./mq/saramakafka
```

## 内置 Hook

### 1. 追踪 Hook

自动添加 OpenTelemetry 追踪：

```go
producer := saramakafka.NewProducer(
    brokers,
    saramakafka.WithTracer(tracer),
)
```

### 2. 耗时统计 Hook

自动记录消息发送耗时：

```go
// 默认已添加，无需手动配置
```

## 配置说明

### 生产者配置

| 配置项 | 默认值 | 说明 |
|-------|-------|------|
| `Version` | V1_0_0_0 | Kafka 版本 |
| `Return.Successes` | true | 返回成功确认 |
| `RequiredAcks` | WaitForAll | 确认机制 |
| `Partitioner` | RoundRobin | 分区策略 |

### 安全配置

```yaml
# SASL 认证
Net.SASL.Enable: true
Net.SASL.User: username
Net.SASL.Password: password

# TLS 加密
Net.TLS.Enable: true
```

## 应用场景

### 1. 事件驱动架构

```go
// 订单创建后发送事件
producer.Publish("order.created", orderID, orderJSON)
```

### 2. 日志收集

```go
// 发送日志到 Kafka
producer.Publish("app.logs", traceID, logJSON)
```

### 3. 数据同步

```go
// 数据库变更同步到 ES
producer.Publish("db.changes", table, changeJSON)
```

### 4. 消息队列

```go
// 异步任务处理
producer.Publish("tasks.email", userID, taskJSON)
```

## 最佳实践

### 1. 生产者单例

```go
// 全局生产者
var producer saramakafka.IProducer

func init() {
    producer = saramakafka.NewProducer(brokers)
}
```

### 2. 错误重试

```go
func sendWithRetry(producer IProducer, topic, key, payload string) error {
    for i := 0; i < 3; i++ {
        if err := producer.Publish(topic, key, payload); err == nil {
            return nil
        }
        time.Sleep(time.Second * time.Duration(i))
    }
    return errors.New("failed to send message")
}
```

### 3. 消息格式

```go
// 使用 JSON 格式
message := map[string]interface{}{
    "event":     "user.created",
    "timestamp": time.Now().Unix(),
    "data":      userData,
}
payload, _ := json.Marshal(message)
producer.Publish("events", "", string(payload))
```

## 注意事项

1. **资源释放**: 使用 `defer producer.Release()` 或 `defer consumer.Close()`
2. **连接管理**: 不要频繁创建/销毁生产者/消费者
3. **消息大小**: 注意 Kafka 消息大小限制（默认 1MB）
4. **Offset 提交**: 消费者确保正确提交 offset
