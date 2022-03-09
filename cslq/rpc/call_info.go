package rpc

import "time"

// CallInfo contains some info about an executed RPC call.
type CallInfo struct {
	Name      string
	Args      []interface{}
	Vals      []interface{}
	Response  int
	CallStart time.Time
	CallEnd   time.Time
}

// Duration returns the execution time of the function. If function was never called, it returns -1.
func (info CallInfo) Duration() time.Duration {
	if info.CallStart.IsZero() || info.CallEnd.IsZero() {
		return -1
	}
	return info.CallEnd.Sub(info.CallStart)
}
