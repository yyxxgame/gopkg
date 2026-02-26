# XTrace - 分布式追踪

基于 OpenTelemetry 的分布式追踪工具。

## 功能

- **OpenTelemetry 集成**: 完整的 OTel 支持
- **多导出器**: 支持 Jaeger、Zipkin、OTLP 等
- **上下文传播**: 自动传递追踪上下文
- **Span 管理**: 简化 Span 创建和管理

## 支持的导出器

- **Jaeger**: UDP/HTTP 导出
- **Zipkin**: HTTP 导出
- **OTLP**: OpenTelemetry Protocol（gRPC/HTTP）
- **Stdout**: 标准输出（调试用）

## 使用示例

### 初始化追踪

```go
package main

import (
    "context"
    "github.com/yyxxgame/gopkg/xtrace"
)

func main() {
    // 初始化 Jaeger 导出器
    err := xtrace.InitJaeger("my-service", "http://jaeger:14268/api/traces")
    if err != nil {
        panic(err)
    }
    defer xtrace.Close()
    
    // 业务逻辑
    doWork(context.Background())
}
```

### 创建 Span

```go
package main

import (
    "context"
    "time"
    "github.com/yyxxgame/gopkg/xtrace"
)

func doWork(ctx context.Context) {
    // 创建 Span
    ctx, span := xtrace.StartSpan(ctx, "doWork")
    defer span.End()
    
    // 添加属性
    span.SetAttribute("user_id", "12345")
    span.SetAttribute("operation", "create_order")
    
    // 业务逻辑
    time.Sleep(100 * time.Millisecond)
    
    // 调用子函数（自动传递上下文）
    processOrder(ctx)
}

func processOrder(ctx context.Context) {
    ctx, span := xtrace.StartSpan(ctx, "processOrder")
    defer span.End()
    
    // 处理订单逻辑
}
```

### 记录错误

```go
func handleError(ctx context.Context, err error) {
    _, span := xtrace.StartSpan(ctx, "handleError")
    defer span.End()
    
    if err != nil {
        // 记录错误
        span.RecordError(err)
        // 设置 Span 状态
        span.SetStatus(codes.Error, err.Error())
    }
}
```

### 添加事件

```go
func processWithEvents(ctx context.Context) {
    ctx, span := xtrace.StartSpan(ctx, "processWithEvents")
    defer span.End()
    
    // 添加事件
    span.AddEvent("start processing")
    
    // 业务逻辑
    time.Sleep(50 * time.Millisecond)
    
    span.AddEvent("step 1 completed")
    
    // 更多逻辑
    span.AddEvent("processing completed")
}
```

## API 参考

### 初始化函数

#### InitJaeger

初始化 Jaeger 导出器。

```go
func InitJaeger(serviceName, endpoint string) error
```

#### InitZipkin

初始化 Zipkin 导出器。

```go
func InitZipkin(serviceName, endpoint string) error
```

#### InitOTLP

初始化 OTLP 导出器。

```go
func InitOTLP(serviceName, endpoint string) error
```

#### InitStdout

初始化标准输出导出器（调试用）。

```go
func InitStdout(serviceName string) error
```

#### Close

关闭追踪器。

```go
func Close()
```

### Span 管理

#### StartSpan

创建新的 Span。

```go
func StartSpan(ctx context.Context, name string) (context.Context, trace.Span)
```

**参数**:
- `ctx`: 父上下文
- `name`: Span 名称

**返回值**:
- 新上下文（包含 Span 上下文）
- Span 实例

### 辅助函数

详细 API 请参考 OpenTelemetry 官方文档。

## 测试

```bash
go test -v ./xtrace
```

## 配置选项

### 采样率

```go
// 设置采样率（0.0 - 1.0）
// 1.0 = 100% 采样
xtrace.InitJaeger("service", endpoint, xtrace.WithSampleRate(1.0))
```

### 资源属性

```go
// 添加资源属性
xtrace.InitJaeger("service", endpoint, 
    xtrace.WithAttributes(
        attribute.String("version", "1.0.0"),
        attribute.String("environment", "production"),
    ))
```

## 与 go-zero 集成

```go
package main

import (
    "github.com/zeromicro/go-zero/core/service"
    "github.com/zeromicro/go-zero/core/trace"
    "github.com/yyxxgame/gopkg/xtrace"
)

func main() {
    // 初始化追踪
    trace.StartAgent(trace.Config{
        Name:     "my-service",
        Endpoint: "http://jaeger:14268/api/traces",
        Sampler:  1.0,
        Batcher:  "jaeger",
    })
    
    // 使用 xtrace 创建 Span
    // ...
    
    // 启动服务
    service.Start()
}
```

## 应用场景

- **微服务追踪**: 追踪请求在多个服务间的流转
- **性能分析**: 识别性能瓶颈
- **错误追踪**: 定位错误发生位置
- **依赖分析**: 分析服务间依赖关系

## 最佳实践

### 1. Span 命名

```go
// ✅ 好的命名
StartSpan(ctx, "UserService.GetUser")
StartSpan(ctx, "OrderService.CreateOrder")
StartSpan(ctx, "DB.Query")

// ❌ 避免的命名
StartSpan(ctx, "process")
StartSpan(ctx, "handle")
```

### 2. 属性添加

```go
// 添加有意义的属性
span.SetAttribute("user.id", userID)
span.SetAttribute("order.amount", amount)
span.SetAttribute("db.table", "users")
```

### 3. 错误处理

```go
if err != nil {
    span.RecordError(err)
    span.SetStatus(codes.Error, err.Error())
    return err
}
```

### 4. 上下文传递

```go
// ✅ 始终传递上下文
func handler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    doWork(ctx) // 传递上下文
}

// ❌ 不要使用 context.Background()
func handler(w http.ResponseWriter, r *http.Request) {
    doWork(context.Background()) // 丢失追踪链路
}
```

## 注意事项

1. **性能影响**: 生产环境建议设置合适的采样率
2. **资源清理**: 程序退出前调用 `Close()`
3. **敏感信息**: 不要在属性中记录敏感数据
4. **Span 数量**: 避免创建过多 Span 影响性能
