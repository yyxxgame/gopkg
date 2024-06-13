//@File     backoff_test.go
//@Time     2024/6/13
//@Author   #Suyghur,

package algorithm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBackoff(t *testing.T) {
	b := NewBackOff()
	dur1 := b.Duration()
	assert.True(t, dur1 > 0)

	dur2 := b.Duration()
	assert.True(t, dur2 >= dur1)
}
