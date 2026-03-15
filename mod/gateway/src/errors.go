package gateway

import "fmt"

type ErrParseError struct {
	msg string
}

func (e ErrParseError) Error() string {
	if len(e.msg) == 0 {
		return "parse error"
	}
	return fmt.Sprintf("parse error: %s", e.msg)
}
