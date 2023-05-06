package node

import (
	_log "github.com/cryptopunkscc/astrald/log"
	"time"
)

var log = _log.Tag("node")

type aliasString string

func init() {
	// set formatters for basic types
	_log.SetFormatter("", func(v interface{}) string {
		return log.EmColor() + v.(string) + log.Reset()
	})

	_log.SetFormatter(aliasString(""), func(v interface{}) string {
		return log.Green() + string(v.(aliasString)) + log.Reset()
	})

	_log.SetFormatter(time.Duration(0), func(i interface{}) string {
		return log.Purple() + i.(time.Duration).String() + log.Reset()
	})

	_log.SetFormatter(nil, func(i interface{}) string {
		return log.Red() + "nil" + log.Reset()
	})
}
