//@File     client_test.go
//@Time     2023/01/12
//@Author   #Suyghur,

package etcdv3

import "testing"

func TestNew(t *testing.T) {
	New([]string{"127.0.0.1:2379"})
}
