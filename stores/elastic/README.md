# Stores/Elastic - Elasticsearch 客户端

Elasticsearch 客户端封装。

## 功能

- **ES 客户端**: 基于 olivere/elastic/v7
- **连接管理**: 自动连接池管理
- **健康检查**: 启动时检查 ES 集群健康状态

## 使用示例

### 基本用法

```go
package main

import (
    "context"
    "fmt"
    "github.com/yyxxgame/gopkg/stores/elastic"
)

func main() {
    // 创建 ES 客户端
    client, err := elastic.NewElastic("http://localhost:9200")
    if err != nil {
        panic(err)
    }
    
    ctx := context.Background()
    
    // 索引文档
    _, err = client.Index().
        Index("users").
        Id("1").
        BodyJson(map[string]interface{}{
            "name": "zhangsan",
            "age":  25,
        }).
        Do(ctx)
    if err != nil {
        panic(err)
    }
    
    // 查询文档
    result, err := client.Get().
        Index("users").
        Id("1").
        Do(ctx)
    if err != nil {
        panic(err)
    }
    fmt.Printf("文档：%+v\n", result.Source)
    
    // 搜索
    searchResult, err := client.Search().
        Index("users").
        Query(elastic.NewMatchQuery("name", "zhangsan")).
        Do(ctx)
    if err != nil {
        panic(err)
    }
    fmt.Printf("找到 %d 条记录\n", searchResult.TotalHits())
}
```

### 集群连接

```go
package main

import (
    "github.com/yyxxgame/gopkg/stores/elastic"
)

func main() {
    // 连接 ES 集群
    client, err := elastic.NewElastic(
        "http://es1:9200",
        "http://es2:9200",
        "http://es3:9200",
    )
    if err != nil {
        panic(err)
    }
    
    // 使用客户端
    // ...
}
```

### 带认证的连接

```go
package main

import (
    "github.com/yyxxgame/gopkg/stores/elastic"
)

func main() {
    client, err := elastic.NewElasticWithAuth(
        "http://localhost:9200",
        "elastic",      // 用户名
        "password",     // 密码
    )
    if err != nil {
        panic(err)
    }
    
    // 使用客户端
    // ...
}
```

## API 参考

### 函数

#### NewElastic

创建 ES 客户端。

```go
func NewElastic(urls ...string) (*elastic.Client, error)
```

**参数**:
- `urls`: ES 集群地址列表

#### NewElasticWithAuth

创建带认证的 ES 客户端。

```go
func NewElasticWithAuth(url, username, password string) (*elastic.Client, error)
```

## 测试

```bash
go test -v ./stores/elastic
```

## 常用操作

### 创建索引

```go
_, err := client.CreateIndex("users").
    BodyJson(map[string]interface{}{
        "mappings": map[string]interface{}{
            "properties": map[string]interface{}{
                "name": map[string]interface{}{
                    "type": "keyword",
                },
                "age": map[string]interface{}{
                    "type": "integer",
                },
                "created_at": map[string]interface{}{
                    "type": "date",
                },
            },
        },
    }).
    Do(ctx)
```

### 删除索引

```go
_, err := client.DeleteIndex("users").Do(ctx)
```

### 批量操作

```go
bulkRequest := client.Bulk()

bulkRequest = bulkRequest.Add(elastic.NewBulkIndexRequest().
    Index("users").
    Id("1").
    Doc(map[string]interface{}{"name": "user1"}))

bulkRequest = bulkRequest.Add(elastic.NewBulkIndexRequest().
    Index("users").
    Id("2").
    Doc(map[string]interface{}{"name": "user2"}))

_, err := bulkRequest.Do(ctx)
```

### 聚合查询

```go
agg := elastic.NewTermsAggregation().Field("age")

searchResult, err := client.Search().
    Index("users").
    Size(0).
    Aggregation("age_groups", agg).
    Do(ctx)
```

## 注意事项

1. **版本兼容**: 基于 olivere/elastic/v7，兼容 ES 7.x
2. **连接池**: 客户端内部自动管理连接池
3. **健康检查**: 初始化时会检查集群健康状态
4. **超时设置**: 建议设置合适的请求超时

## 性能优化

### 1. 使用 Bulk API

```go
// 批量插入，减少网络往返
bulkRequest := client.Bulk()
for i := 0; i < 1000; i++ {
    bulkRequest = bulkRequest.Add(elastic.NewBulkIndexRequest().
        Index("logs").
        Doc(logData))
}
bulkRequest.Do(ctx)
```

### 2. 使用 Scroll 查询大量数据

```go
scroll := client.Scroll("large-index").
    Size(1000).
    Query(query)

for {
    result, err := scroll.Do(ctx)
    if err != nil {
        break
    }
    
    // 处理结果
    for _, hit := range result.Hits.Hits {
        // ...
    }
    
    if len(result.Hits.Hits) == 0 {
        break
    }
}

scroll.Clear(ctx)
```

### 3. 使用 Source Filtering

```go
// 只返回需要的字段
result, err := client.Search().
    Index("users").
    Source(elastic.NewSearchSource().
        FetchSource([]string{"name", "email"}, nil)).
    Do(ctx)
```
