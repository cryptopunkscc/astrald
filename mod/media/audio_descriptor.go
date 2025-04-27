package media

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/object"
	"github.com/cryptopunkscc/astrald/streams"
	"io"
)

var _ astral.Object = &AudioDescriptor{}

type AudioDescriptor struct {
	Format  string
	Title   string
	Artist  string
	Album   string
	Genre   string
	Year    int
	Picture *object.ID
}

func (d *AudioDescriptor) WriteTo(w io.Writer) (n int64, err error) {
	c := streams.NewWriteCounter(w)
	err = cslq.Encode(c, "[c]c [c]c [c]c [c]c [c]c s v",
		d.Format, d.Title, d.Artist, d.Album, d.Genre, d.Year, &d.Picture)
	n = c.Total()
	return
}

func (d *AudioDescriptor) ReadFrom(r io.Reader) (n int64, err error) {
	c := streams.NewReadCounter(r)
	err = cslq.Decode(c, "[c]c [c]c [c]c [c]c [c]c s v",
		&d.Format, &d.Title, &d.Artist, &d.Album, &d.Genre, &d.Year, &d.Picture)
	n = c.Total()
	return
}

func (AudioDescriptor) ObjectType() string { return "astrald.mod.media.audio_descriptor" }

func (d *AudioDescriptor) String() string {
	s := d.Title
	if s == "" {
		s = "Untitled"
	}
	if d.Artist != "" {
		s = s + " by " + d.Artist
	}
	return s
}
