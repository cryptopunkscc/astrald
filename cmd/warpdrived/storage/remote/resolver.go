package remote

import (
	"github.com/cryptopunkscc/astrald/cmd/warpdrived/storage"
	"github.com/cryptopunkscc/astrald/proto/android/content"
	"github.com/cryptopunkscc/astrald/proto/warpdrive"
	"path"
)

var _ storage.FileResolver = Resolver{}

type Resolver struct {
	content.Client
}

func (c Resolver) Info(uri string) (files []warpdrive.Info, err error) {
	i, err := c.Client.Info(uri)
	if err != nil {
		return
	}
	files = append(files, warpdrive.Info{
		Uri:   i.Uri,
		Size:  i.Size,
		Mime:  i.Mime,
		Name:  resolveName(i),
		IsDir: false,
		Perm:  0755,
	})
	return
}

func resolveName(i content.Info) string {
	switch {
	case i.Name != "":
		return i.Name
	default:
		return path.Base(i.Uri)
	}
}
