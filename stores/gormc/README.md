# gormc

Wrapper by gorm and cache of go-zero

## Create you model file

```go
package model

import (
	"context"
	"fmt"
	"github.com/yyxxgame/gopkg/stores/gormc"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"gorm.io/gorm"
)

var (
	//Define your cache key prefix in here.
	cacheKeyPrefix = fmt.Sprintf("%s%s", gormc.CachePrefix, "your_key:kf:")
)

type (
	// ITestModel Define your model export interface here what you want
	ITestModel interface {
		FindOneByUsername(ctx context.Context, username string) (*UserInfo, error)
	}

	UserInfo struct {
		Id       int64  `gorm:"column:id;type:int(11);primaryKey;not null;" json:"id"`
		Username string `gorm:"column:username;type:varchar(50);not null;" json:"username"`
	}

	// testModel Define your model
	testModel struct {
		gormc.IGormc
	}
)

// TableName Impl Tabler interface of gormc to change table name what you want.
func (info *UserInfo) TableName() string {
	return "t_users"
}

func NewTestModel(conn *gorm.DB, c cache.CacheConf, opts ...gormc.Option) ITestModel {
	return testModel{
		IGormc: gormc.NewConn(conn, c, opts...),
	}
}

func (m *testModel) FindOneByUsername(ctx context.Context, username string) (*UserInfo, error) {
	//TODO implement your logic
	panic("implement your logic")
}
```

## Use your model

```go
package dao

import (
	"github.com/yyxxgame/gopkg/stores/gormc"
	"github.com/zeromicro/go-zero/core/stores/cache"
	oteltrace "go.opentelemetry.io/otel/trace"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"time"
)

func NewDaoMapper(dsn string, cacheConf cache.CacheConf, tracer oteltrace.Tracer) {
	// gormc logger
	gormLogger := gormc.NewGormLogger("dev")
	// create gorm db
	if db, err := gorm.Open(mysql.Open("dsl"), &gorm.Config{Logger: gormLogger}); err != nil {
		panic(err)
	} else {
		// use gormc metric plguin
		_ = db.Use(&gormc.MetricPlugin{})

		testModel := model.NewTestModel(db, cacheConf, gormc.WithTracer(tracer), gormc.WithExpiry(time.Hour*12))
	}
}

```