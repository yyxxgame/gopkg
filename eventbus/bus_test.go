//@File     bus_test.go
//@Time     2023/01/11
//@Author   #Suyghur,

package eventbus

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestWatch(t *testing.T) {
	eBus := NewEventBus()
	var triggered bool
	var lock sync.RWMutex
	eBus.Watch("trigger", func(event *Event) {
		lock.Lock()
		triggered = true
		lock.Unlock()
	})
	lock.RLock()
	assert.Equal(t, !triggered, true)
	lock.RUnlock()

	eBus.Dispatch(&Event{Name: "not-trigger"})
	time.Sleep(time.Millisecond)
	lock.RLock()
	assert.Equal(t, !triggered, true)
	lock.RUnlock()

	eBus.Dispatch(&Event{Name: "trigger"})
	time.Sleep(time.Millisecond)
	lock.RLock()
	assert.Equal(t, triggered, true)
	lock.RUnlock()
}

func TestUnWatch(t *testing.T) {
	eBus := NewEventBus()
	var triggered bool
	var notTriggered bool
	var lock sync.RWMutex
	triggeredHandler := func(event *Event) {
		lock.Lock()
		triggered = true
		lock.Unlock()
	}
	notTriggeredHandler := func(event *Event) {
		lock.Lock()
		notTriggered = true
		lock.Unlock()
	}

	eBus.Watch("trigger", triggeredHandler)
	eBus.Watch("trigger", notTriggeredHandler)
	eBus.Unwatch("trigger", notTriggeredHandler)

	eBus.Dispatch(&Event{Name: "trigger"})
	time.Sleep(time.Millisecond)
	lock.RLock()
	assert.Equal(t, triggered, true)
	assert.Equal(t, !notTriggered, true)
	lock.RUnlock()
}

func TestDispatchAndWait(t *testing.T) {
	eBus := NewEventBus()

	var triggered bool
	var mu sync.RWMutex
	delayHandler := func(event *Event) {
		time.Sleep(time.Second)
		mu.Lock()
		triggered = true
		mu.Unlock()
	}

	eBus.Watch("trigger", delayHandler)

	eBus.DispatchAndWait(&Event{Name: "trigger"})
	mu.RLock()
	assert.Equal(t, triggered, true)
	mu.RUnlock()
}
