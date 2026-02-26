# MQ/Tasks - 任务队列核心

基于 go-queue 的任务队列核心实现。

## 功能

- **任务路由**: 支持任务路由配置
- **Kafka 消费**: 从 Kafka 消费任务消息
- **追踪集成**: 支持 OpenTelemetry 追踪
- **任务工厂**: 工厂模式创建任务

## 接口定义

### ITaskServerCtx

任务服务器上下文接口。

```go
type ITaskServerCtx interface {
    TaskRoute() map[string]string      // 任务路由
    KafkaConf() []kq.KqConf           // Kafka 配置
    Telemetry() trace.Config          // 追踪配置
    FuncRegister() FuncRegister       // 任务注册函数
    Object() interface{}              // 服务对象
}
```

### ITaskFactory

任务工厂接口。

```go
type ITaskFactory interface {
    Gen(key string, t FuncCreateTask)
}
```

### ITask

任务接口。

```go
type ITask interface {
    Run(ctx context.Context, k, v string) error
    Stop()
}
```

## 使用示例

### 定义任务

```go
package task

import (
    "context"
    "github.com/yyxxgame/gopkg/mq/tasks/core"
)

type ProcessOrderTask struct {
    // 任务数据
}

func NewProcessOrderTask() core.ITask {
    return &ProcessOrderTask{}
}

func (t *ProcessOrderTask) Run(ctx context.Context, key, value string) error {
    // 处理订单逻辑
    // ...
    return nil
}

func (t *ProcessOrderTask) Stop() {
    // 清理资源
}
```

### 配置任务路由

```go
type TaskServer struct {
    taskRoute map[string]string
    kafkaConf []kq.KqConf
}

func (s *TaskServer) TaskRoute() map[string]string {
    return map[string]string{
        "order": "order-tasks",    // 任务键 → Kafka 主题
        "payment": "payment-tasks",
    }
}

func (s *TaskServer) KafkaConf() []kq.KqConf {
    return []kq.KqConf{
        {
            Brokers: []string{"localhost:9092"},
            Group:   "task-consumer",
            Topic:   "order-tasks",
        },
    }
}
```

### 注册任务

```go
func main() {
    // 创建任务工厂
    factory := tasks.NewTaskFactory(serverCtx)
    
    // 注册任务
    factory.Gen("order", task.NewProcessOrderTask)
    factory.Gen("payment", task.NewPaymentTask)
    
    // 启动服务
    // ...
}
```

## 测试

```bash
go test -v ./mq/tasks
```

## 任务路由配置

任务路由将任务键映射到 Kafka 主题：

```go
TaskRoute: {
    "email":     "email-tasks",
    "sms":       "sms-tasks",
    "notification": "notification-tasks",
}
```

## 追踪集成

```go
func (s *TaskServer) Telemetry() trace.Config {
    return trace.Config{
        Name:     "task-server",
        Endpoint: "http://jaeger:14268/api/traces",
        Sampler:  1.0,
        Batcher:  "jaeger",
    }
}
```

## 最佳实践

### 1. 任务分类

按业务类型分类任务：

```go
TaskRoute: {
    "order.create":  "order-create-tasks",
    "order.cancel":  "order-cancel-tasks",
    "payment.refund": "payment-refund-tasks",
}
```

### 2. 错误隔离

```go
func (t *Task) Run(ctx context.Context, key, value string) error {
    defer func() {
        if r := recover(); r != nil {
            logx.ErrorStack(r)
        }
    }()
    
    return t.process(key, value)
}
```

### 3. 监控指标

```go
func (t *Task) Run(ctx context.Context, key, value string) error {
    startTime := time.Now()
    defer func() {
        duration := time.Since(startTime)
        metrics.TaskDuration.Observe(duration.Seconds())
    }()
    
    return t.process(key, value)
}
```

## 注意事项

1. **任务路由**: 确保路由配置正确
2. **消费者组**: 每个主题使用独立的消费者组
3. **幂等性**: 任务需要支持重复执行
4. **资源清理**: 在 `Stop()` 中释放资源
