package objects

import (
	"fmt"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	log2 "github.com/cryptopunkscc/astrald/mod/log"
)

type RepositoryInfoView struct {
	*RepositoryInfo
}

func (v RepositoryInfoView) Render() string {
	var size = astral.Size(v.Free)

	return fmt.Sprintf("%s: %s (%s free)",
		log2.GrayText.Render(string(v.Name)),
		log2.WhiteText.Render(string(v.Label)),
		log2.SizeView{Size: &size}.Render(),
	)
}

func init() {
	log.DefaultViewer.Set(RepositoryInfo{}.ObjectType(), func(object astral.Object) astral.Object {
		return &RepositoryInfoView{object.(*RepositoryInfo)}
	})
}
