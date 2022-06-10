package block

import (
	"errors"
)

const (
	cmdRead     = 0x01
	cmdWrite    = 0x02
	cmdSeek     = 0x03
	cmdFinalize = 0x04
	cmdClose    = 0xff
)

const (
	success        = 0x00
	errEOB         = 0x01
	errNoSpace     = 0x02
	errFailed      = 0xfe
	errUnavailable = 0xff
)

const (
	SeekStart   = 0
	SeekCurrent = 1
	SeekEnd     = 2
)

var ErrFailed = errors.New("failed")
var ErrUnavailable = errors.New("unavailable")
var ErrEnded = errors.New("protocol ended")

type ProtocolError struct {
	e string
}

func (e ProtocolError) Error() string {
	return "protocol error: " + e.e
}
