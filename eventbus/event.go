//@File     event.go
//@Time     2023/01/11
//@Author   #Suyghur,

package eventbus

import "time"

type Event struct {
	Name    string
	Time    time.Time
	Payload string
	Extra   interface{}
}
