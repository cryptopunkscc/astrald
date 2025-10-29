package frames

import (
	"fmt"
	"io"
)

const MaxPayloadSize = 8 * 1024 // 8kb

type Frame interface {
	io.ReaderFrom
	io.WriterTo
	fmt.Stringer
}

const (
	opPing = iota
	opQuery
	opResponse
	opRead
	opData
	opReset
	opMigrate
)
