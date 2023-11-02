//@File     cached_gorm_test.go
//@Time     2023/10/23
//@Author   #Suyghur,

package gormc

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/yyxxgame/gopkg/internal/dbtest"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"regexp"
	"sync/atomic"
	"testing"
	"time"
)

type (
	trackedGormc struct {
		redis *redis.Redis
		IGormc
		execValue      bool
		queryRowsValue bool
		transactValue  bool
	}
)

func mockGormc(t *testing.T, db *sql.DB) *trackedGormc {
	c := &trackedGormc{}

	c.redis = dbtest.CreateRedis(t)
	assert.NotNil(t, c.redis)

	gormDb, err := gorm.Open(mysql.New(mysql.Config{
		SkipInitializeWithVersion: true,
		Conn:                      db,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("an error '%s' was not expected when initialize a gorm instance", err)
	}
	c.IGormc = NewNodeConn(gormDb, c.redis, WithExpiry(time.Second*10))

	return c
}

func resetStats() {
	atomic.StoreUint64(&stats.Total, 0)
	atomic.StoreUint64(&stats.Hit, 0)
	atomic.StoreUint64(&stats.Miss, 0)
	atomic.StoreUint64(&stats.DbFails, 0)
}

func TestGormc_SetCache(t *testing.T) {
	dbtest.RunTest(t, func(db *sql.DB, mock sqlmock.Sqlmock) {
		c := mockGormc(t, db)

		assert.Nil(t, c.SetCache("any", "value"))
		value, _ := c.redis.Get("any")
		assert.Equal(t, `"value"`, value)
	})
}

func TestGormc_GetCache(t *testing.T) {
	dbtest.RunTest(t, func(db *sql.DB, mock sqlmock.Sqlmock) {
		c := mockGormc(t, db)

		var value string
		assert.Equal(t, ErrNotFound, c.GetCache("any", &value))

		assert.NoError(t, c.redis.Set("any", `"value"`))

		assert.Nil(t, c.GetCache("any", &value))
		assert.Equal(t, "value", value)
	})
}

func TestGormc_DelCache(t *testing.T) {
	dbtest.RunTest(t, func(db *sql.DB, mock sqlmock.Sqlmock) {
		c := mockGormc(t, db)

		assert.NoError(t, c.redis.Set("any", `"value"`))

		assert.Nil(t, c.DelCache("any"))

		value, err := c.redis.Get("any")
		assert.Nil(t, err)
		assert.Empty(t, value)
	})
}

func TestGormc_Query(t *testing.T) {
	dbtest.RunTest(t, func(db *sql.DB, mock sqlmock.Sqlmock) {
		c := mockGormc(t, db)
		resp := map[string]interface{}{}
		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `t_users` WHERE `t_users`.`id` = ? LIMIT 1")).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "suyghur"))

		err := c.Query(&resp, "any", func(conn *gorm.DB, value interface{}) error {
			c.queryRowsValue = true
			return conn.Table("`t_users`").Where("`t_users`.`id` = ?", 1).Take(&resp).Error
		})
		assert.Nil(t, err)
		assert.True(t, c.queryRowsValue)

		bResp, _ := json.Marshal(resp)
		value, _ := c.redis.Get("any")
		assert.Equal(t, bResp, []byte(value))
	})
}

func TestGormc_QueryNoCache(t *testing.T) {
	dbtest.RunTest(t, func(db *sql.DB, mock sqlmock.Sqlmock) {
		c := mockGormc(t, db)
		resp := map[string]interface{}{}
		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `t_users` WHERE `t_users`.`id` = ? LIMIT 1")).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "suyghur"))

		err := c.QueryNoCache(&resp, func(conn *gorm.DB, value interface{}) error {
			c.queryRowsValue = true
			return conn.Table("`t_users`").Where("`t_users`.`id` = ?", 1).Take(&resp).Error
		})
		assert.Nil(t, err)
		assert.True(t, c.queryRowsValue)
	})
}

func TestGormc_QueryWithExpire(t *testing.T) {
	dbtest.RunTest(t, func(db *sql.DB, mock sqlmock.Sqlmock) {
		c := mockGormc(t, db)
		resp := map[string]interface{}{}
		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `t_users` WHERE `t_users`.`id` = ? LIMIT 1")).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "suyghur"))

		err := c.QueryWithExpire(&resp, "any", time.Second*30, func(conn *gorm.DB, value interface{}) error {
			c.queryRowsValue = true
			return conn.Table("`t_users`").Where("`t_users`.`id` = ?", 1).Take(&resp).Error
		})
		assert.Nil(t, err)
		assert.True(t, c.queryRowsValue)

		bResp, _ := json.Marshal(resp)
		value, _ := c.redis.Get("any")
		ttl, _ := c.redis.Ttl("any")
		assert.Equal(t, bResp, []byte(value))
		assert.True(t, ttl > 0 && time.Duration(ttl) <= time.Minute)
	})
}

func TestGormc_QueryRowIndex(t *testing.T) {
	dbtest.RunTest(t, func(db *sql.DB, mock sqlmock.Sqlmock) {
		c := mockGormc(t, db)

		assert.Nil(t, c.redis.Set("index", `"primary"`))

		var resp string
		err := c.QueryRowIndex(&resp, "index", func(primary interface{}) string {
			return fmt.Sprintf("%s/1234", primary)
		}, func(conn *gorm.DB, value interface{}) (interface{}, error) {
			assert.Fail(t, "should not go here")
			return "primary", nil
		}, func(conn *gorm.DB, value, primary interface{}) error {
			*value.(*string) = "xin"
			assert.Equal(t, "primary", primary)
			return nil
		})

		assert.Nil(t, err)
		assert.Equal(t, "xin", resp)

		value, err := c.redis.Get("index")
		assert.Nil(t, err)
		assert.Equal(t, `"primary"`, value)
		value, err = c.redis.Get("primary/1234")
		assert.Nil(t, err)
		assert.Equal(t, `"xin"`, value)
	})
}

func TestGormc_Exec(t *testing.T) {
	dbtest.RunTest(t, func(db *sql.DB, mock sqlmock.Sqlmock) {
		c := mockGormc(t, db)

		mock.ExpectExec("DELETE FROM `t_users` WHERE `t_users`.`id` = ?").WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 1))

		err := c.Exec(func(conn *gorm.DB) error {
			c.execValue = true
			return conn.Exec("DELETE FROM `t_users` WHERE `t_users`.`id` = ?", 1).Error
		})
		assert.Nil(t, err)
		assert.True(t, c.execValue)
	})
}

func TestGormc_Transact(t *testing.T) {
	dbtest.RunTest(t, func(db *sql.DB, mock sqlmock.Sqlmock) {
		c := mockGormc(t, db)

		mock.ExpectBegin()
		mock.ExpectExec("UPDATE `t_users` SET `t_users`.`name` = ?").WithArgs("suyghur").WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec("INSERT INTO `t_users`").WithArgs("yyxxgame").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := c.Transact(func(tx *gorm.DB) error {
			if err := tx.Exec("UPDATE `t_users` SET `t_users`.`name` = ?", "suyghur").Error; err != nil {
				return err
			}

			if err := tx.Exec("INSERT INTO `t_users`", "yyxxgame").Error; err != nil {
				return err
			}

			c.transactValue = true
			return nil
		})
		assert.Nil(t, err)
		assert.True(t, c.transactValue)
	})
}
