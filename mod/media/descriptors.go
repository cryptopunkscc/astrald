package media

type Desc Info

func (Desc) Type() string {
	return "mod.media.info"
}
func (d Desc) String() string {
	s := d.Title
	if d.Artist != "" {
		s = d.Artist + " - " + s
	}
	return s
}
