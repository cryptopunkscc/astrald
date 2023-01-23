package bt

import _log "github.com/cryptopunkscc/astrald/log"

var log = _log.Tag(NetworkName)

func init() {
	_log.SetFormatter(Addr{}, func(i interface{}) string {
		addr, _ := i.(Addr)
		return log.Sformat("%s%s%s", log.Cyan(), addr.String(), log.Reset())
	})
}
