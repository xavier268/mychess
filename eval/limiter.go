package eval

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

type Limit struct {
	ctx   context.Context
	limit float64 // max heap space percentage accepted
}

var ErrLimitExceeded = fmt.Errorf("heap space limit exceeded")

func NewLimit(ctx context.Context, limit float64) *Limit {
	return &Limit{ctx: ctx, limit: limit}
}

// Default to 15 sec time out and 90% heap space
func NewDefaultLimit() *Limit {
	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	return NewLimit(ctx, 0.98)
}

func (l *Limit) Check() error {
	if err := l.ctx.Err(); err != nil {
		return l.ctx.Err()
	}
	if HeapPercentage() >= l.limit {
		return ErrLimitExceeded
	}
	return nil
}

func HeapPercentage() float64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return float64(m.HeapAlloc) / float64(m.HeapSys)
}

func HeapValue() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.HeapAlloc
}
