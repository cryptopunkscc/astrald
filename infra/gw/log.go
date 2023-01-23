package gw

import _log "github.com/cryptopunkscc/astrald/log"

var log = _log.Tag("gw")

func init() {
	_log.SetFormatter(Addr{}, func(i interface{}) string {
		addr, _ := i.(Addr)
		return log.Sformat("%s:%s", addr.gate, addr.target)
	})
}
