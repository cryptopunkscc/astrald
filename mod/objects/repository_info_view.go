package objects

import (
	"fmt"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/log/styles"
	"github.com/cryptopunkscc/astrald/mod/log/views"
)

type RepositoryInfoView struct {
	*RepositoryInfo
}

func (v RepositoryInfoView) Render() string {
	var size = astral.Size(v.Free)

	return fmt.Sprintf("%s: %s (%s free)",
		styles.GrayText.Render(string(v.Name)),
		styles.WhiteText.Render(string(v.Label)),
		views.SizeView{Size: &size}.Render(),
	)
}

func init() {
	log.DefaultViewer.Set(RepositoryInfo{}.ObjectType(), func(object astral.Object) astral.Object {
		return &RepositoryInfoView{object.(*RepositoryInfo)}
	})
}
