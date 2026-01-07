package media

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	log2 "github.com/cryptopunkscc/astrald/mod/log"
)

type AudioFileView struct {
	*AudioFile
}

func (view AudioFileView) Render() string {
	return log.Render(log.Format("%v by %v (%v)",
		log2.String(string(view.Title), &log2.BrightGreenText),
		log2.String(string(view.Artist), &log2.BrightGreenText),
		log2.String(string(view.Album), &log2.GreenText),
	)...)
}

func init() {
	log.DefaultViewer.Set(AudioFile{}.ObjectType(), func(object astral.Object) astral.Object {
		return &AudioFileView{object.(*AudioFile)}
	})
}
