package warpdrived

import (
	"github.com/cryptopunkscc/astrald/lib/warpdrived/core"
	"path/filepath"
)

func Android(dir string) *Server {
	return &Server{
		debug: true,
		Component: core.Component{
			Config: core.Config{
				Platform:      core.PlatformAndroid,
				RepositoryDir: filepath.Join(dir, "warpdrive"),
			},
		},
	}
}
