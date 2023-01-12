//@File     queue_test.go
//@Time     2023/01/11
//@Author   #Suyghur,

package eventbus

import (
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestQueue(t *testing.T) {
	size := 10
	q := NewQueue(size)
	for i := 0; i < size; i++ {
		n := "E." + strconv.Itoa(i)
		q.Push(&Event{Name: n})
		es := q.Dump().([]*Event)
		assert.Equal(t, len(es) == i+1, true)
		assert.Equal(t, es[0].Name == n, true)
	}
	for i := 0; i < size; i++ {
		q.Push(&Event{Name: strconv.Itoa(i)})
		es := q.Dump().([]*Event)
		assert.Equal(t, len(es) == size, true)
	}
}

func TestQueueInvalidCapacity(t *testing.T) {
	defer func() {
		e := recover()
		assert.Equal(t, e != nil, true)
	}()
	NewQueue(-1)
}

func BenchmarkQueue(b *testing.B) {
	q := NewQueue(100)
	e := &Event{}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Push(e)
	}
}

func BenchmarkQueueConcurrent(b *testing.B) {
	q := NewQueue(100)
	e := &Event{}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			q.Push(e)
		}
	})
	b.StopTimer()
}
