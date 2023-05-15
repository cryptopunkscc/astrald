package node

import (
	"github.com/cryptopunkscc/astrald/log"
	"time"
)

type aliasString string

func init() {
	// set formatters for basic types
	log.SetFormatter("", func(v interface{}) string {
		return log.EmColor() + v.(string) + log.Reset()
	})

	log.SetFormatter(aliasString(""), func(v interface{}) string {
		return log.Green() + string(v.(aliasString)) + log.Reset()
	})

	log.SetFormatter(time.Duration(0), func(i interface{}) string {
		return log.Purple() + i.(time.Duration).String() + log.Reset()
	})

	log.SetFormatter(nil, func(i interface{}) string {
		return log.Red() + "nil" + log.Reset()
	})
}
