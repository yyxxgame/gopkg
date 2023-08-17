//@Auther   : Kaishin
//@Time     : 2022/12/22

package watcher

import "testing"

var (
	endpoints = []string{"localhost:2379"}
	key       = "foo"
	value     = "hello, world"
)

func TestWatcher_AddListener(t *testing.T) {
	w, err := NewWatcher(endpoints, key)
	if err != nil {
		t.Fatal(err)
	}

	w.AddListener(func(k, v string) {
		t.Logf("event: %s, value: %v", k, v)
	})

	t.Log("Listening...")
	select {}
}

func TestWatcher_Notify(t *testing.T) {
	w, err := NewWatcher(endpoints, key)
	if err != nil {
		t.Fatal(err)
	}

	err = w.Notify(value)
	if err != nil {
		t.Fatal(err)
	}
}
