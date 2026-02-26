# Queue - 任务队列 v1

基于 go-queue 的任务队列框架。

## 功能

- **任务注册**: 支持注册多个任务处理器
- **Kafka 消费**: 从 Kafka 消费任务消息
- **服务集成**: 与 go-zero 服务组集成
- **任务工厂**: 支持工厂模式创建任务

## 接口定义

### IService

任务服务接口。

```go
type IService interface {
    service.Service
    Object() interface{}
    Gen(key string, t FuncTask)
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
    "encoding/json"
    "github.com/yyxxgame/gopkg/infrastructure/queue"
)

type EmailTask struct {
    // 任务数据
}

func NewEmailTask() queue.ITask {
    return &EmailTask{}
}

func (t *EmailTask) Run(ctx context.Context, key, value string) error {
    // 解析任务数据
    var data map[string]interface{}
    if err := json.Unmarshal([]byte(value), &data); err != nil {
        return err
    }
    
    // 执行发送邮件逻辑
    // ...
    
    return nil
}

func (t *EmailTask) Stop() {
    // 清理资源
}
```

### 注册任务

```go
package main

import (
    "github.com/zeromicro/go-zero/core/service"
    "github.com/yyxxgame/gopkg/infrastructure/queue"
    "your-project/task"
)

func main() {
    // 创建任务服务
    taskService := queue.NewTaskService(conf)
    
    // 注册任务工厂
    taskService.Gen("email", task.NewEmailTask)
    
    // 添加到服务组
    group := service.NewServiceGroup()
    group.Add(taskService)
    
    group.Start()
}
```

## 测试

```bash
go test -v ./infrastructure/queue
```

## 任务生命周期

1. **创建**: 通过工厂函数创建任务实例
2. **运行**: 调用 `Run()` 处理消息
3. **停止**: 服务关闭时调用 `Stop()`

## 错误处理

```go
func (t *EmailTask) Run(ctx context.Context, key, value string) error {
    // 处理失败返回错误，消息会重新入队
    if err := process(value); err != nil {
        return err
    }
    return nil
}
```

## 与 v2 的区别

| 特性 | v1 | v2 |
|------|----|----|
| 接口设计 | 简单 | 更完善 |
| 配置方式 | 基础 | 支持更多选项 |
| 推荐使用 | ❌ 维护模式 | ✅ 推荐使用 |

## 注意事项

1. **幂等性**: 任务需要支持重复执行
2. **资源清理**: 在 `Stop()` 中释放资源
3. **错误处理**: 返回错误会导致消息重试
