package bt

import "log"

const enableDebug = false

func debugln(v ...interface{}) {
	if enableDebug {
		log.Println(v...)
	}
}

func debugf(fmt string, v ...interface{}) {
	if enableDebug {
		log.Printf(fmt, v...)
	}
}
