package udp

import "time"

// segMeta stores metadata for an unacked segment
type segMeta struct {
	data     []byte
	sentAt   time.Time
	retries  int
	seqStart uint32
	length   int
}
