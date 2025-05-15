package gateway

import (
	"errors"
	"fmt"
)

var ErrInvalidGateway = errors.New("invalid gateway")
var ErrAlreadySubscribed = errors.New("already subscribed")
var ErrNotSubscribed = errors.New("subscription not found")

type ErrParseError struct {
	msg string
}

func (e ErrParseError) Error() string {
	if len(e.msg) == 0 {
		return "parse error"
	}
	return fmt.Sprintf("parse error: %s", e.msg)
}
