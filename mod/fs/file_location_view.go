package fs

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/log/styles"
	"github.com/cryptopunkscc/astrald/mod/log/views"
)

type FileLocationView struct {
	*FileLocation
}

func (v *FileLocationView) Render() string {
	return log.DefaultViewer.Render(log.Format(
		"file at %v:%v",
		v.NodeID,
		views.String(string(v.Path), &styles.DarkYellowText),
	)...)
}

func init() {
	log.DefaultViewer.Set(FileLocation{}.ObjectType(), func(object astral.Object) astral.Object {
		return &FileLocationView{object.(*FileLocation)}
	})
}
