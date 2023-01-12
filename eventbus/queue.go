//@File     queue.go
//@Time     2023/01/11
//@Author   #Suyghur,

package eventbus

import (
	"sync"
	"sync/atomic"
)

// MaxEventNum is the default size of a event queue.
const MaxEventNum = 200

type (
	queue struct {
		ring        []*Event
		tail        uint32
		tailVersion map[uint32]*uint32
		lock        sync.RWMutex
	}
	IQueue interface {
		Push(e *Event)
		Dump() interface{}
	}
)

// NewQueue creates a queue with the given capacity.
func NewQueue(cap int) IQueue {
	q := &queue{
		ring:        make([]*Event, cap),
		tailVersion: make(map[uint32]*uint32, cap),
	}
	for i := 0; i < cap; i++ {
		t := uint32(0)
		q.tailVersion[uint32(i)] = &t
	}
	return q
}

// Push pushes an event to the queue.
func (q *queue) Push(e *Event) {
	for {
		oldTail := atomic.LoadUint32(&q.tail)
		newTail := oldTail + 1
		if newTail >= uint32(len(q.ring)) {
			newTail = 0
		}
		oldVersion := atomic.LoadUint32(q.tailVersion[oldTail])
		newVersion := oldVersion + 1
		if atomic.CompareAndSwapUint32(&q.tail, oldTail, newTail) && atomic.CompareAndSwapUint32(q.tailVersion[oldTail], oldVersion, newVersion) {
			q.lock.Lock()
			q.ring[oldTail] = e
			q.lock.Unlock()
			break
		}
	}
}

// Dump dumps the previously pushed events out in a reversed order.
func (q *queue) Dump() interface{} {
	results := make([]*Event, 0, len(q.ring))
	q.lock.RLock()
	defer q.lock.RUnlock()
	pos := int32(q.tail)
	for i := 0; i < len(q.ring); i++ {
		pos--
		if pos < 0 {
			pos = int32(len(q.ring) - 1)
		}
		e := q.ring[pos]
		if e == nil {
			return results
		}
		results = append(results, e)
	}
	return results
}
