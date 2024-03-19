//@File     concurrent_map_test.go
//@Time     2024/03/19
//@Author   #Suyghur,

package collection

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"reflect"
	"sync"
	"testing"
)

func TestConcurrentMap_Set_Get(t *testing.T) {
	cm := NewConcurrentMap[string, int](100)

	var wg1 sync.WaitGroup
	wg1.Add(10)

	for i := 0; i < 10; i++ {
		go func(n int) {
			cm.Set(fmt.Sprintf("%d", n), n)
			wg1.Done()
		}(i)
	}
	wg1.Wait()

	for j := 0; j < 10; j++ {
		go func(n int) {
			val, ok := cm.Get(fmt.Sprintf("%d", n))
			assert.Equal(t, n, val)
			assert.Equal(t, true, ok)
		}(j)
	}
}

func TestConcurrentMap_GetOrSet(t *testing.T) {
	cm := NewConcurrentMap[string, int](100)

	var wg sync.WaitGroup
	wg.Add(5)

	for i := 0; i < 5; i++ {
		go func(n int) {
			val, ok := cm.GetOrSet(fmt.Sprintf("%d", n), n)
			assert.Equal(t, n, val)
			assert.Equal(t, false, ok)
			wg.Done()
		}(i)
	}
	wg.Wait()

	for j := 0; j < 5; j++ {
		go func(n int) {
			val, ok := cm.Get(fmt.Sprintf("%d", n))
			assert.Equal(t, n, val)
			assert.Equal(t, true, ok)
		}(j)
	}
}

func TestConcurrentMap_Delete(t *testing.T) {
	cm := NewConcurrentMap[string, int](100)

	var wg1 sync.WaitGroup
	wg1.Add(10)

	for i := 0; i < 10; i++ {
		go func(n int) {
			cm.Set(fmt.Sprintf("%d", n), n)
			wg1.Done()
		}(i)
	}
	wg1.Wait()

	var wg2 sync.WaitGroup
	wg2.Add(10)

	for i := 0; i < 10; i++ {
		go func(n int) {
			cm.Delete(fmt.Sprintf("%d", n))
			wg2.Done()
		}(i)
	}
	wg2.Wait()

	for j := 0; j < 10; j++ {
		go func(n int) {
			_, ok := cm.Get(fmt.Sprintf("%d", n))
			assert.Equal(t, false, ok)
		}(j)
	}
}

func TestConcurrentMap_GetAndDelete(t *testing.T) {
	cm := NewConcurrentMap[string, int](100)

	var wg1 sync.WaitGroup
	wg1.Add(10)
	for i := 0; i < 10; i++ {
		go func(n int) {
			cm.Set(fmt.Sprintf("%d", n), n)
			wg1.Done()
		}(i)
	}
	wg1.Wait()

	var wg2 sync.WaitGroup
	wg2.Add(10)
	for j := 0; j < 10; j++ {
		go func(n int) {
			val, ok := cm.GetAndDelete(fmt.Sprintf("%d", n))
			assert.Equal(t, n, val)
			assert.Equal(t, true, ok)

			_, ok = cm.Get(fmt.Sprintf("%d", n))
			assert.Equal(t, false, ok)
			wg2.Done()
		}(j)
	}

	wg2.Wait()
}

func TestConcurrentMap_Has(t *testing.T) {
	cm := NewConcurrentMap[string, int](100)

	var wg1 sync.WaitGroup
	wg1.Add(10)

	for i := 0; i < 10; i++ {
		go func(n int) {
			cm.Set(fmt.Sprintf("%d", n), n)
			wg1.Done()
		}(i)
	}
	wg1.Wait()

	for j := 0; j < 10; j++ {
		go func(n int) {
			ok := cm.Has(fmt.Sprintf("%d", n))
			assert.Equal(t, true, ok)
		}(j)
	}
}

func TestConcurrentMap_Range(t *testing.T) {
	cm := NewConcurrentMap[string, int](100)

	var wg1 sync.WaitGroup
	wg1.Add(10)

	for i := 0; i < 10; i++ {
		go func(n int) {
			cm.Set(fmt.Sprintf("%d", n), n)
			wg1.Done()
		}(i)
	}
	wg1.Wait()

	cm.Range(func(key string, value int) bool {
		assert.Equal(t, key, fmt.Sprintf("%d", value))
		return true
	})
}

func TestConcurrentMap_Keys(t *testing.T) {
	cm := NewConcurrentMap[int, string](100)

	cm.Set(1, "one")
	cm.Set(2, "two")
	cm.Set(3, "three")

	keys := cm.Keys()

	expectedKeys := []int{3, 2, 1}
	eq := reflect.DeepEqual(keys, expectedKeys)
	assert.Equal(t, true, eq)
}

func TestConcurrentMap_Values(t *testing.T) {
	cm := NewConcurrentMap[int, string](100)

	cm.Set(1, "one")
	cm.Set(2, "two")
	cm.Set(3, "three")

	values := cm.Values()

	expectedValues := []string{"three", "two", "one"}
	eq := reflect.DeepEqual(values, expectedValues)
	assert.Equal(t, true, eq)
}
