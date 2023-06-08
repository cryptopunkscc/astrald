package log

import (
	"time"
)

type Printer interface {
	Log(t Type, level int, ts time.Time, tag string, ops ...Op)
}
