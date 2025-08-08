//@File     period_limit.go
//@Time     2025/8/7
//@Author   #Suyghur,

package redis

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"strconv"
	"time"

	v9rds "github.com/redis/go-redis/v9"
)

const (
	// Unknown means not initialized state.
	Unknown = iota
	// Allowed means allowed state.
	Allowed
	// HitQuota means this request exactly hit the quota.
	HitQuota
	// OverQuota means passed the quota.
	OverQuota

	internalOverQuota = 0
	internalAllowed   = 1
	internalHitQuota  = 2
)

var (
	// ErrUnknownCode is an error that represents unknown status code.
	ErrUnknownCode = errors.New("unknown status code")

	//go:embed internal/periodscript.lua
	periodLuaScript string
	periodScript    = v9rds.NewScript(periodLuaScript)
)

type (
	// PeriodOption defines the method to customize a PeriodLimit.
	PeriodOption func(l *PeriodLimit)

	// A PeriodLimit is used to limit requests during a period of time.
	PeriodLimit struct {
		period     int
		quota      int
		limitStore *Redis
		keyPrefix  string
		align      bool
	}
)

// NewPeriodLimit returns a PeriodLimit with given parameters.
func NewPeriodLimit(period, quota int, limitStore *Redis, keyPrefix string, opts ...PeriodOption) *PeriodLimit {
	limiter := &PeriodLimit{
		period:     period,
		quota:      quota,
		limitStore: limitStore,
		keyPrefix:  keyPrefix,
	}

	for _, opt := range opts {
		opt(limiter)
	}

	return limiter
}

// Take requests a permit, it returns the permit state.
func (p *PeriodLimit) Take(key string) (int, error) {
	return p.TakeCtx(context.Background(), key)
}

// TakeCtx requests a permit with context, it returns the permit state.
func (p *PeriodLimit) TakeCtx(ctx context.Context, key string) (int, error) {
	resp, err := periodScript.Run(ctx, p.limitStore,
		[]string{fmt.Sprintf("%s%s", p.keyPrefix, key)},
		[]string{strconv.Itoa(p.quota), strconv.Itoa(p.calcExpireSeconds())},
	).Result()
	if err != nil {
		return Unknown, err
	}

	code, ok := resp.(int64)
	if !ok {
		return Unknown, ErrUnknownCode
	}

	switch code {
	case internalOverQuota:
		return OverQuota, nil
	case internalAllowed:
		return Allowed, nil
	case internalHitQuota:
		return HitQuota, nil
	default:
		return Unknown, ErrUnknownCode
	}
}

func (p *PeriodLimit) calcExpireSeconds() int {
	if p.align {
		now := time.Now()
		_, offset := now.Zone()
		unix := now.Unix() + int64(offset)
		return p.period - int(unix%int64(p.period))
	}

	return p.period
}

// Align returns a func to customize a PeriodLimit with alignment.
// For example, if we want to limit end users with 5 sms verification messages every day,
// we need to align with the local timezone and the start of the day.
func Align() PeriodOption {
	return func(l *PeriodLimit) {
		l.align = true
	}
}
