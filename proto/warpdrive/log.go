package warpdrive

import (
	"fmt"
	"log"
	"os"
)

func NewLogger(prefix ...interface{}) *log.Logger {
	var chunks []interface{}
	suffix := " "
	for i, chunk := range prefix {
		if i == len(prefix)-1 {
			suffix = " < "
		}
		chunks = append(chunks, fmt.Sprint(chunk)+suffix)
	}
	return log.New(os.Stderr, fmt.Sprint(chunks...), log.LstdFlags|log.Lmsgprefix)
}
