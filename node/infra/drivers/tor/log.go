package tor

import (
	"fmt"
	_log "github.com/cryptopunkscc/astrald/log"
)

var log = _log.Tag(DriverName)

func init() {
	_log.SetFormatter(Endpoint{}, func(i interface{}) string {
		e, _ := i.(Endpoint)
		return fmt.Sprintf("%s%s%s", log.Cyan(), e.String(), log.Reset())
	})
}
