package service

import (
	"fmt"
	"log"
	"os"
)

func (srv *Context) LogPrefix(prefix ...string) {
	logger := NewLogger(prefix...)
	srv.Logger = logger
	srv.SetLogger(logger)
}

func NewLogger(prefix ...string) *log.Logger {
	var chunks []interface{}
	suffix := " "
	for i, chunk := range prefix {
		if i == len(prefix)-1 {
			suffix = ": "
		}
		chunks = append(chunks, chunk+suffix)
	}
	return log.New(os.Stderr, fmt.Sprint(chunks...), log.LstdFlags|log.Lmsgprefix)
}
