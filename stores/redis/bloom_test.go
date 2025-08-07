//@File     bloom_test.go
//@Time     2025/5/29
//@Author   #Suyghur,

package redis

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBitset(t *testing.T) {
	rds := CreateRedis(t)
	ctx := context.Background()

	bitSet := newBitset(rds, "test_key", 1024)
	isSetBefore, err := bitSet.check(ctx, []uint{0})
	if err != nil {
		t.Fatal(err)
	}
	if isSetBefore {
		t.Fatal("Bit should not be set")
	}
	err = bitSet.set(ctx, []uint{512})
	if err != nil {
		t.Fatal(err)
	}
	isSetAfter, err := bitSet.check(ctx, []uint{512})
	if err != nil {
		t.Fatal(err)
	}
	if !isSetAfter {
		t.Fatal("Bit should be set")
	}
	err = bitSet.expire(3600)
	if err != nil {
		t.Fatal(err)
	}
	err = bitSet.del()
	if err != nil {
		t.Fatal(err)
	}
}

func TestBloomAdd(t *testing.T) {
	rds := CreateRedis(t)
	b := NewBloomFilter(rds, "test_key", 64)
	assert.Nil(t, b.Add([]byte("hello")))
	assert.Nil(t, b.Add([]byte("world")))
	ok, err := b.Exists([]byte("hello"))
	assert.Nil(t, err)
	assert.True(t, ok)
}

func TestBloomExists(t *testing.T) {
	rds, clean := CreateRedisWithCloseHandler(t)
	b := NewBloomFilter(rds, "test_key", 64)
	_, err := b.Exists([]byte{0, 1, 2})
	assert.NoError(t, err)

	clean()
	b = NewBloomFilter(rds, "test", 64)
	_, err = b.Exists([]byte{0, 1, 2})
	assert.Error(t, err)
}
