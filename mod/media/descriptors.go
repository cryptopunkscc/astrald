package media

type Descriptor Info

func (Descriptor) DescriptorType() string {
	return "mod.media.info"
}
func (d Descriptor) String() string {
	s := d.Title
	if d.Artist != "" {
		s = d.Artist + " - " + s
	}
	return s
}
