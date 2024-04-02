package media

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/lib/desc"
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

// Video descriptor
type Video struct {
	Format   string
	Title    string
	Duration time.Duration
}

var _ desc.Data = &Video{}

func (*Video) Type() string { return "mod.media.video" }
func (v *Video) String() string {
	if len(v.Title) > 0 {
		return v.Title
	}

	return "Untitled video"
}

// Image descriptor
type Image struct {
	Format string
	Width  int
	Height int
}

var _ desc.Data = &Image{}

func (*Image) Type() string { return "mod.media.image" }
func (i *Image) String() string {
	return fmt.Sprintf("%dx%dpx image", i.Width, i.Height)

}
