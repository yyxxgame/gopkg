//@File     bus.go
//@Time     2023/01/11
//@Author   #Suyghur,

package eventbus

import (
	"context"
	"github.com/yyxxgame/gopkg/xsync/gopool"
	"reflect"
	"sync"
)

type (
	// Callback is called when the subscribed event happens.
	Callback func(*Event)
	// IBus implements the observer pattern to allow event dispatching and watching.
	IBus interface {
		Watch(event string, callback Callback)
		Unwatch(event string, callback Callback)
		Dispatch(event *Event)
		DispatchAndWait(event *Event)
	}
	bus struct {
		callbacks sync.Map
	}
)

// NewEventBus creates a Bus.
func NewEventBus() IBus {
	return &bus{}
}

// Watch subscribe to a certain event with a callback.
func (b *bus) Watch(event string, callback Callback) {
	var callbacks []Callback
	if actual, ok := b.callbacks.Load(event); ok {
		callbacks = append(callbacks, actual.([]Callback)...)
	}
	callbacks = append(callbacks, callback)
	b.callbacks.Store(event, callbacks)
}

// Unwatch remove the given callback from the callback chain of an event.
func (b *bus) Unwatch(event string, callback Callback) {
	var filtered []Callback
	// In go, functions are not comparable, so we use reflect.ValueOf(callback).Pointer() to reflect their address for comparison.
	target := reflect.ValueOf(callback).Pointer()
	if actual, ok := b.callbacks.Load(event); ok {
		for _, h := range actual.([]Callback) {
			if reflect.ValueOf(h).Pointer() != target {
				filtered = append(filtered, h)
			}
		}
		b.callbacks.Store(event, filtered)
	}
}

// Dispatch dispatches an event by invoking each callback asynchronously.
func (b *bus) Dispatch(event *Event) {
	if actual, ok := b.callbacks.Load(event.Name); ok {
		for _, h := range actual.([]Callback) {
			f := h // assign the value to a new variable for the closure
			gopool.CtxGo(context.Background(), func() {
				f(event)
			})
		}
	}
}

// DispatchAndWait dispatches an event by invoking callbacks concurrently and waits for them to finish.
func (b *bus) DispatchAndWait(event *Event) {
	if actual, ok := b.callbacks.Load(event.Name); ok {
		var wg sync.WaitGroup
		for i := range actual.([]Callback) {
			h := (actual.([]Callback))[i]
			wg.Add(1)
			gopool.CtxGo(context.Background(), func() {
				h(event)
				wg.Done()
			})
		}
		wg.Wait()
	}
}
