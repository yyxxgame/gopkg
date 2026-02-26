# EventBus - 事件总线

基于观察者模式实现的事件总线，支持异步和同步事件分发。

## 功能

- **事件订阅**: 订阅感兴趣的事件
- **事件取消订阅**: 取消对事件的订阅
- **异步事件分发**: 非阻塞方式分发事件
- **同步事件分发**: 阻塞方式等待所有回调执行完成

## 使用示例

### 基本用法

```go
package main

import (
    "fmt"
    "github.com/yyxxgame/gopkg/eventbus"
)

func main() {
    // 创建事件总线
    bus := eventbus.NewEventBus()
    
    // 订阅事件
    bus.Watch("user.created", func(event *eventbus.Event) {
        fmt.Printf("用户创建：%s\n", event.Payload)
    })
    
    bus.Watch("user.created", func(event *eventbus.Event) {
        fmt.Printf("发送欢迎邮件：%s\n", event.Payload)
    })
    
    // 创建事件
    event := &eventbus.Event{
        Name:    "user.created",
        Payload: "用户 ID: 12345",
        Extra:   map[string]interface{}{"userId": 12345},
    }
    
    // 异步分发事件（非阻塞）
    bus.Dispatch(event)
    
    // 同步分发事件（阻塞，等待所有回调完成）
    bus.DispatchAndWait(event)
    
    // 取消订阅
    // 注意：需要保持回调函数引用才能取消订阅
}
```

### 带额外数据的事件

```go
package main

import (
    "fmt"
    "github.com/yyxxgame/gopkg/eventbus"
)

type UserEvent struct {
    UserID   int64
    Username string
    Email    string
}

func main() {
    bus := eventbus.NewEventBus()
    
    // 订阅用户事件
    bus.Watch("user.registered", func(event *eventbus.Event) {
        userEvent := event.Extra.(*UserEvent)
        fmt.Printf("新用户注册：%s (%s)\n", userEvent.Username, userEvent.Email)
    })
    
    // 分发带结构化数据的事件
    bus.Dispatch(&eventbus.Event{
        Name: "user.registered",
        Extra: &UserEvent{
            UserID:   1001,
            Username: "zhangsan",
            Email:    "zhangsan@example.com",
        },
    })
}
```

## API 参考

### 类型

#### Event

事件结构体。

```go
type Event struct {
    Name    string      // 事件名称
    Time    time.Time   // 事件时间
    Payload string      // 事件负载（字符串）
    Extra   interface{} // 额外数据（任意类型）
}
```

#### Callback

回调函数类型。

```go
type Callback func(*Event)
```

#### IBus

事件总线接口。

```go
type IBus interface {
    Watch(event string, callback Callback)
    Unwatch(event string, callback Callback)
    Dispatch(event *Event)
    DispatchAndWait(event *Event)
}
```

### 函数

#### NewEventBus

创建新的事件总线。

```go
func NewEventBus() IBus
```

#### Watch

订阅事件。

```go
func (b *bus) Watch(event string, callback Callback)
```

**参数**:
- `event`: 事件名称
- `callback`: 回调函数

#### Unwatch

取消订阅事件。

```go
func (b *bus) Unwatch(event string, callback Callback)
```

**参数**:
- `event`: 事件名称
- `callback`: 要移除的回调函数

#### Dispatch

异步分发事件（非阻塞）。

```go
func (b *bus) Dispatch(event *Event)
```

**特点**:
- 每个回调在独立的 goroutine 中执行
- 立即返回，不等待回调完成

#### DispatchAndWait

同步分发事件（阻塞）。

```go
func (b *bus) DispatchAndWait(event *Event)
```

**特点**:
- 每个回调在独立的 goroutine 中执行
- 等待所有回调完成后返回

## 测试

```bash
go test -v ./eventbus
```

## 实现细节

- **并发安全**: 使用 `sync.Map` 存储回调
- **异步执行**: 使用 goroutine 池（gopool）执行回调
- **函数比较**: 使用反射比较函数指针来移除回调

## 应用场景

- **解耦模块**: 模块间通过事件通信，降低耦合
- **审计日志**: 记录系统中的重要事件
- **通知系统**: 邮件、短信、推送等通知
- **数据同步**: 数据变更后的同步操作
- **插件系统**: 支持插件扩展功能

## 最佳实践

1. **事件命名**: 使用有意义的名称，如 `user.created`、`order.paid`
2. **错误处理**: 在回调中处理错误，避免 panic
3. **性能考虑**: 异步分发不会阻塞主流程
4. **避免循环**: 不要在回调中分发相同事件，避免无限循环
