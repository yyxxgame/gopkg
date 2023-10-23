//@File     metric_test.go
//@Time     2023/10/23
//@Author   #Suyghur,

package gormc

import (
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/yyxxgame/gopkg/internal/dbtest"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"testing"
)

func TestUseMetricPlugin(t *testing.T) {
	dbtest.RunTest(t, func(db *sql.DB, mock sqlmock.Sqlmock) {
		gormDB, err := gorm.Open(mysql.New(mysql.Config{
			SkipInitializeWithVersion: true,
			Conn:                      db,
		}), &gorm.Config{})
		assert.Nil(t, err)

		err = gormDB.Use(&MetricPlugin{})
		assert.Nil(t, err)
	})
}
