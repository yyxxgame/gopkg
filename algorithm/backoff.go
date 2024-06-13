//@File     backoff.go
//@Time     2024/6/13
//@Author   #Suyghur,

package algorithm

import (
	"math"
	"sync/atomic"
	"time"
)

type (
	Backoff struct {
		attempts atomic.Int64
		// 乘数因子
		factor float64
		// 扰动因子
		jitter    float64
		baseDelay time.Duration
		maxDelay  time.Duration
	}

	BackoffOption func(*Backoff)
)

func NewBackOff(opts ...BackoffOption) *Backoff {
	b := &Backoff{
		factor:    2.0,
		jitter:    1,
		baseDelay: time.Second,
		maxDelay:  120 * time.Second,
	}

	for _, option := range opts {
		option(b)
	}

	return b
}

func (b *Backoff) Duration() time.Duration {
	// jitter * baseDelay * factor^N(attempts)
	dur := b.baseDelay * time.Duration(b.jitter*math.Pow(b.factor, float64(b.attempts.Load())))
	if dur > b.maxDelay {
		return b.maxDelay
	}

	b.attempts.Add(1)

	return dur
}

func (b *Backoff) Reset() {
	b.attempts.Store(0)
}

func (b *Backoff) Attempts() int64 {
	return b.attempts.Load()
}

func WithBaseDelay(baseDelay time.Duration) BackoffOption {
	return func(b *Backoff) {
		b.baseDelay = baseDelay
	}
}

func WithMaxDelay(maxDelay time.Duration) BackoffOption {
	return func(b *Backoff) {
		b.maxDelay = maxDelay
	}
}

func WithFactor(factor float64) BackoffOption {
	return func(b *Backoff) {
		b.factor = factor
	}
}

func WithJitter(jitter float64) BackoffOption {
	return func(b *Backoff) {
		b.jitter = jitter
	}
}
