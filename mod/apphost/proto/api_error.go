package proto

import "github.com/cryptopunkscc/astrald/cslq"

type APIError interface {
	Code() int
	Error() string
}

type apiError struct {
	code  int
	error string
}

var apiErrors = map[int]*apiError{}

func makeError(code int, msg string) *apiError {
	apiErrors[code] = &apiError{
		code:  code,
		error: msg,
	}

	return apiErrors[code]
}

func (e apiError) MarshalCSLQ(enc *cslq.Encoder) error {
	return enc.Encodef("c", e.code)
}

func (e apiError) Code() int {
	return e.code
}

func (e apiError) Error() string {
	return e.error
}

func ErrorCode(code int) error {
	if code == Success {
		return nil
	}
	if err, found := apiErrors[code]; found {
		return err
	}

	return ErrUnknown
}
