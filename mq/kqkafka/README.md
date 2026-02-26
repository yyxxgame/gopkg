# MQ/KqKafka - Kafka Q 消费者

基于 go-queue 的 Kafka 消费者实现。

## 功能

- **消费者组**: 支持 Kafka 消费者组
- **消息处理**: 支持自定义消息处理器
- **自动重连**: 连接断开自动重连
- **优雅关闭**: 支持优雅关闭

## 使用示例

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "github.com/yyxxgame/gopkg/mq/kqkafka"
)

type Message struct {
    ID      int64  `json:"id"`
    Content string `json:"content"`
}

func main() {
    // 创建消费者配置
    conf := kqkafka.KqConf{
        Brokers: []string{"localhost:9092"},
        Group:   "my-consumer-group",
        Topic:   "my-topic",
    }
    
    // 创建消费者
    consumer := kqkafka.NewKafkaConsumer(conf)
    
    // 消费消息
    ctx := context.Background()
    err := consumer.Consume(ctx, func(ctx context.Context, key, value string) error {
        var msg Message
        if err := json.Unmarshal([]byte(value), &msg); err != nil {
            return err
        }
        
        fmt.Printf("收到消息：%+v\n", msg)
        
        // 处理业务逻辑
        // ...
        
        return nil
    })
    
    if err != nil {
        panic(err)
    }
}
```

## API 参考

### 类型

#### KqConf

消费者配置。

```go
type KqConf struct {
    Brokers []string `json:"Brokers"` // Kafka  brokers
    Group   string   `json:"Group"`   // 消费者组
    Topic   string   `json:"Topic"`   // 主题
}
```

#### Consumer

消费者接口。

```go
type Consumer interface {
    Consume(ctx context.Context, handler func(context.Context, string, string) error) error
}
```

### 函数

#### NewKafkaConsumer

创建 Kafka 消费者。

```go
func NewKafkaConsumer(conf KqConf) Consumer
```

## 测试

```bash
go test -v ./mq/kqkafka
```

## 消息处理

### 错误处理

```go
consumer.Consume(ctx, func(ctx context.Context, key, value string) error {
    // 处理消息
    if err := processMessage(key, value); err != nil {
        // 返回错误会导致消息重新入队
        return err
    }
    return nil
})
```

### 批量处理

```go
var batch []string
batchSize := 100

consumer.Consume(ctx, func(ctx context.Context, key, value string) error {
    batch = append(batch, value)
    
    if len(batch) >= batchSize {
        // 批量处理
        processBatch(batch)
        batch = batch[:0]
    }
    
    return nil
})
```

## 应用场景

### 1. 异步任务处理

```go
// 接收任务消息
consumer.Consume(ctx, func(ctx context.Context, key, value string) error {
    var task Task
    json.Unmarshal([]byte(value), &task)
    
    // 执行任务
    executeTask(task)
    
    return nil
})
```

### 2. 数据同步

```go
// 同步数据库变更
consumer.Consume(ctx, func(ctx context.Context, key, value string) error {
    var change DataChange
    json.Unmarshal([]byte(value), &change)
    
    // 同步到 ES/Redis
    syncToES(change)
    
    return nil
})
```

### 3. 事件驱动

```go
// 处理领域事件
consumer.Consume(ctx, func(ctx context.Context, key, value string) error {
    var event DomainEvent
    json.Unmarshal([]byte(value), &event)
    
    // 触发业务逻辑
    handleEvent(event)
    
    return nil
})
```

## 注意事项

1. **消费者组**: 确保正确配置消费者组实现负载均衡
2. **Offset 提交**: 消息处理成功后自动提交 offset
3. **重复消费**: 业务逻辑需要支持幂等性
4. **优雅关闭**: 使用 context 控制关闭
