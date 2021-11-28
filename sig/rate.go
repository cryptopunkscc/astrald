package sig

import (
	"context"
	"time"
)

func Rate(ctx context.Context, rate int, burst int) Signal {
	return Pace(ctx, time.Duration(time.Second.Nanoseconds()/int64(rate)), burst)
}
