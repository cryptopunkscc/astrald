package media

import (
	"github.com/cryptopunkscc/astrald/lib/desc"
	"github.com/cryptopunkscc/astrald/object"
	"time"
)

// Audio descriptor
type Audio struct {
	Format   string
	Duration time.Duration
	Title    string
	Artist   string
	Album    string
	Genre    string
	Year     int
	Picture  object.ID
}

var _ desc.Data = &Audio{}

func (*Audio) Type() string { return "mod.media.audio" }
func (a *Audio) String() string {
	s := a.Title
	if s == "" {
		s = "Untitled"
	}
	if a.Artist != "" {
		s = s + " by " + a.Artist
	}
	return s
}
