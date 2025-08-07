//@File     bloom_filter.go
//@Time     2025/5/29
//@Author   #Suyghur,

package redis

import (
	"context"
	_ "embed"
	"errors"
	"strconv"
	"time"

	v9rds "github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/hash"
)

// for detailed error rate table, see http://pages.cs.wisc.edu/~cao/papers/summary-cache/node8.html
// maps as k in the error rate table
const maps = 14

var (
	// ErrTooLargeOffset indicates the offset is too large in bitset.
	ErrTooLargeOffset = errors.New("too large offset")

	//go:embed internal/bloom_setscript.lua
	bloomSetLuaScript string
	bloomSetScript    = v9rds.NewScript(bloomSetLuaScript)

	//go:embed internal/bloom_testscript.lua
	bloomTestLuaScript string
	bloomTestScript    = v9rds.NewScript(bloomTestLuaScript)
)

type (
	IBloomFilter interface {
		Add(data []byte) error
		AddCtx(ctx context.Context, data []byte) error
		Exists(data []byte) (bool, error)
		ExistsCtx(ctx context.Context, data []byte) (bool, error)
	}
	// A bloomFilter is a bloom bloomFilter.
	bloomFilter struct {
		bits   uint
		bitSet bitSetProvider
	}

	bitSetProvider interface {
		check(ctx context.Context, offsets []uint) (bool, error)
		set(ctx context.Context, offsets []uint) error
	}

	bitset struct {
		rds  *Redis
		key  string
		bits uint
	}
)

// NewBloomFilter create a bloomFilter, store is the backed redis, key is the key for the bloom bloomFilter,
// bits is how many bits will be used, maps is how many hashes for each addition.
// best practices:
// elements - means how many actual elements
// when maps = 14, formula: 0.7*(bits/maps), bits = 20*elements, the error rate is 0.000067 < 1e-4
// for detailed error rate table, see http://pages.cs.wisc.edu/~cao/papers/summary-cache/node8.html
func NewBloomFilter(rds *Redis, key string, bits uint) IBloomFilter {
	return &bloomFilter{
		bits:   bits,
		bitSet: newBitset(rds, key, bits),
	}
}

// Add adds data into f.
func (f *bloomFilter) Add(data []byte) error {
	return f.AddCtx(context.Background(), data)
}

// AddCtx adds data into f with context.
func (f *bloomFilter) AddCtx(ctx context.Context, data []byte) error {
	locations := f.getLocations(data)
	return f.bitSet.set(ctx, locations)
}

// Exists checks if data is in f.
func (f *bloomFilter) Exists(data []byte) (bool, error) {
	return f.ExistsCtx(context.Background(), data)
}

// ExistsCtx checks if data is in f with context.
func (f *bloomFilter) ExistsCtx(ctx context.Context, data []byte) (bool, error) {
	locations := f.getLocations(data)
	isSet, err := f.bitSet.check(ctx, locations)
	if err != nil {
		return false, err
	}

	return isSet, nil
}

func (f *bloomFilter) getLocations(data []byte) []uint {
	locations := make([]uint, maps)
	for i := uint(0); i < maps; i++ {
		hashValue := hash.Hash(append(data, byte(i)))
		locations[i] = uint(hashValue % uint64(f.bits))
	}

	return locations
}

func newBitset(rds *Redis, key string, bits uint) *bitset {
	return &bitset{
		rds:  rds,
		key:  key,
		bits: bits,
	}
}

func (r *bitset) buildOffsetArgs(offsets []uint) ([]string, error) {
	args := make([]string, 0, len(offsets))

	for _, offset := range offsets {
		if offset >= r.bits {
			return nil, ErrTooLargeOffset
		}

		args = append(args, strconv.FormatUint(uint64(offset), 10))
	}

	return args, nil
}

func (r *bitset) check(ctx context.Context, offsets []uint) (bool, error) {
	args, err := r.buildOffsetArgs(offsets)
	if err != nil {
		return false, err
	}

	resp, err := bloomTestScript.Run(ctx, r.rds, []string{r.key}, args).Result()
	if errors.Is(err, v9rds.Nil) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	exists, ok := resp.(int64)
	if !ok {
		return false, nil
	}

	return exists == 1, nil
}

// del only use for testing.
func (r *bitset) del() error {
	return r.rds.Del(context.Background(), r.key).Err()
}

// expire only use for testing.
func (r *bitset) expire(seconds int) error {
	return r.rds.Expire(context.Background(), r.key, time.Duration(seconds)*time.Second).Err()
}

func (r *bitset) set(ctx context.Context, offsets []uint) error {
	args, err := r.buildOffsetArgs(offsets)
	if err != nil {
		return err
	}

	err = bloomSetScript.Run(ctx, r.rds, []string{r.key}, args).Err()
	if errors.Is(err, v9rds.Nil) {
		return nil
	}

	return err
}
