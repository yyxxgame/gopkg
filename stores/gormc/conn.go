//@File     gormc.go
//@Time     2024/6/5
//@Author   #Suyghur,

package gormc

import "gorm.io/gorm"

type (
	IGormConn struct {
	}

	GormConn struct {
		*gorm.DB
	}
)

func NewGormConn() *GormConn {
	conn := &GormConn{}
	return conn
}
