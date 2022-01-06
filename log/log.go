package log

import "log"

const (
	Normal = iota
	Verbose
	Debug
)

func Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func Println(v ...interface{}) {
	log.Println(v...)
}
