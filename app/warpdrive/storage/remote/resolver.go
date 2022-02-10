package remote

import (
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	content "github.com/cryptopunkscc/astrald/mobile/android/service/content/go"
)

var _ api.FileResolver = Resolver{}

type Resolver struct {
	content.Client
}

func (c Resolver) Info(uri string) (files []api.Info, err error) {
	i, err := c.Client.Info(uri)
	if err != nil {
		return
	}
	files = append(files, api.Info{
		Uri:   i.Uri,
		Size:  i.Size,
		Mime:  i.Mime,
		IsDir: false,
		Perm:  0755,
	})
	return
}
