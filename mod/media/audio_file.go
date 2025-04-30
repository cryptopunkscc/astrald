package media

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/object"
	"io"
)

type AudioFile struct {
	ObjectID  *object.ID
	Format    astral.String8
	Title     astral.String8
	Artist    astral.String8
	Album     astral.String8
	Genre     astral.String8
	Year      astral.Uint16
	PictureID *object.ID
}

// astral

var _ astral.Object = &AudioFile{}

func (AudioFile) ObjectType() string { return "mod.media.audio_file" }

func (d AudioFile) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(d).WriteTo(w)
}

func (d *AudioFile) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(d).ReadFrom(r)
}

// json works by default

// text

func (d AudioFile) MarshalText() (text []byte, err error) {
	s := fmt.Sprintf("%s by %s", d.Title, d.Artist)
	return []byte(s), nil
}

func init() {
	astral.DefaultBlueprints.Add(&AudioFile{})
}
