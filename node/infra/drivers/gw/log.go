package gw

import _log "github.com/cryptopunkscc/astrald/log"

var log = _log.Tag("gw")

func init() {
	_log.SetFormatter(Endpoint{}, func(i interface{}) string {
		e, _ := i.(Endpoint)
		return log.Sformat("%s:%s", e.gate, e.target)
	})
}
