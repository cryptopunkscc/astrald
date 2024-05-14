package content

import (
	"context"
	"fmt"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/mod/media"
	"github.com/cryptopunkscc/astrald/object"
)

func (mod *Module) BestTitle(objectID object.ID) string {
	descs := mod.objects.Describe(context.Background(), objectID, desc.DefaultOpts())

	var m = map[string]*desc.Desc{}
	for _, d := range descs {
		t := d.Data.Type()
		cur, found := m[t]
		if !found {
			m[t] = d
			continue
		}
		if mod.compareTrust(cur, d) > 0 {
			m[t] = d
		}
	}

	if d, found := m[content.LabelDesc{}.Type()]; found {
		d, _ := d.Data.(content.LabelDesc)
		return d.Label
	}

	if d, found := m[keys.KeyDesc{}.Type()]; found {
		d, _ := d.Data.(keys.KeyDesc)
		return d.String()
	}

	if d, found := m[fs.FileDesc{}.Type()]; found {
		d, _ := d.Data.(fs.FileDesc)
		return d.String()
	}

	if d, found := m[(&media.Audio{}).Type()]; found {
		d, _ := d.Data.(*media.Audio)
		return d.String()
	}

	if d, found := m[content.TypeDesc{}.Type()]; found {
		d, _ := d.Data.(content.TypeDesc)
		return "Untitled " + d.String()
	}

	for _, d := range descs {
		if s, ok := d.Data.(fmt.Stringer); ok {
			return s.String()
		}
	}

	return objectID.String()
}

func (mod *Module) compareTrust(a, b *desc.Desc) int {
	if a.Source.IsEqual(b.Source) {
		return 0
	}

	return -1
}
