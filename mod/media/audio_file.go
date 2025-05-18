package media

import (
	"encoding/json"
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"io"
)

type AudioFile struct {
	ObjectID  *astral.ObjectID
	Format    astral.String8
	Title     astral.String8
	Artist    astral.String8
	Album     astral.String8
	Genre     astral.String8
	Year      astral.Uint16
	PictureID *astral.ObjectID
}

var _ astral.Object = &AudioFile{}

// astral

func (AudioFile) ObjectType() string { return "mod.media.audio_file" }

func (f AudioFile) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(f).WriteTo(w)
}

func (f *AudioFile) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(f).ReadFrom(r)
}

func (f AudioFile) MarshalText() (text []byte, err error) {
	s := fmt.Sprintf("%s by %s", f.Title, f.Artist)
	return []byte(s), nil
}

// json

func (f AudioFile) MarshalJSON() ([]byte, error) {
	type alias AudioFile
	return json.Marshal(alias(f))
}

func (f *AudioFile) UnmarshalJSON(bytes []byte) error {
	type alias AudioFile
	var a alias

	err := json.Unmarshal(bytes, &a)
	if err != nil {
		return err
	}

	*f = AudioFile(a)
	return nil
}

// text

func init() {
	astral.DefaultBlueprints.Add(&AudioFile{})
}
