package fs

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	log2 "github.com/cryptopunkscc/astrald/mod/log"
)

type FileLocationView struct {
	*FileLocation
}

func (v *FileLocationView) Render() string {
	return log.DefaultViewer.Render(log.Format(
		"file at %v:%v",
		v.NodeID,
		log2.String(string(v.Path), &log2.DarkYellowText),
	)...)
}

func init() {
	log.DefaultViewer.Set(FileLocation{}.ObjectType(), func(object astral.Object) astral.Object {
		return &FileLocationView{object.(*FileLocation)}
	})
}
