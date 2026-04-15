package media

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/log/styles"
	"github.com/cryptopunkscc/astrald/mod/log/views"
)

type AudioFileView struct {
	*AudioFile
}

func (view AudioFileView) Render() string {
	return log.Render(log.Format("%v by %v (%v)",
		views.String(string(view.Title), &styles.BrightGreenText),
		views.String(string(view.Artist), &styles.BrightGreenText),
		views.String(string(view.Album), &styles.GreenText),
	)...)
}

func init() {
	log.DefaultViewer.Set(AudioFile{}.ObjectType(), func(object astral.Object) astral.Object {
		return &AudioFileView{object.(*AudioFile)}
	})
}
