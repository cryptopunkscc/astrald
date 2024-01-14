package gateway

import (
	"errors"
	"fmt"
)

var ErrSelfGateway = errors.New("cannot use self as gateway")
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
