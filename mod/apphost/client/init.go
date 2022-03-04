package astral

import (
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"os"
	"path/filepath"
	"strconv"
)

func init() {
	var (
		target = "tcp:127.0.0.1:" + strconv.Itoa(apphost.TCPPort)
		path   = make([]string, 0)
	)

	if home, err := os.UserConfigDir(); err == nil {
		path = append(path, filepath.Join(home, "astrald", apphost.UnixSocketName))
	}

	path = append(path, filepath.Join(apphost.DefaultSocketDir, apphost.UnixSocketName))

	for _, p := range path {
		if _, err := os.Stat(p); err == nil {
			target = "unix:" + p
			break
		}
	}

	instance = NewAppHost(target)
}
