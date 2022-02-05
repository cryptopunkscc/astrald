package remote

import (
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	content "github.com/cryptopunkscc/astrald/mobile/android/service/content/api"
)

var _ api.FileResolver = Resolver{}

type Resolver struct {
	content.Client
}

func (c Resolver) Info(uri string) (files []api.Info, err error) {
	info, err := c.Client.Info(uri)
	if err != nil {
		return
	}
	for _, i := range info {
		files = append(files, api.Info{
			Uri:   i.Uri,
			Size:  i.Size,
			IsDir: false,
			Perm:  0755,
			Mime:  i.Mime,
		})
	}
	return
}
