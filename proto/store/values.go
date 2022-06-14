package store

import (
	"errors"
)

const (
	cmdOpen     = 0x01
	cmdCreate   = 0x02
	cmdDownload = 0x03
	cmdEnd      = 0xff
)

const (
	success        = 0x00
	errNotFound    = 0x01
	errNoSpace     = 0x02
	errFailed      = 0xfe
	errUnavailable = 0xff
)

var errEnded = errors.New("ended")
