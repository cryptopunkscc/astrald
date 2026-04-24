package media

import (
	"github.com/cryptopunkscc/astrald/astral/fmt"
	"github.com/cryptopunkscc/astrald/mod/log/styles"
	"github.com/cryptopunkscc/astrald/mod/log/theme"
	"github.com/cryptopunkscc/astrald/mod/media"
)

type AudioFileView struct {
	*media.AudioFile
}

func (view AudioFileView) Render() string {
	return fmt.Sprintf("%v by %v (%v)",
		styles.String(string(view.Title), theme.Primary),
		styles.String(string(view.Artist), theme.Secondary),
		styles.String(string(view.Album), theme.Tertiary),
	)
}

func init() {
	fmt.SetView(func(o *media.AudioFile) fmt.View {
		return &AudioFileView{AudioFile: o}
	})
}
