package content

import (
	"context"
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/mod/media"
	"github.com/cryptopunkscc/astrald/object"
)

func (mod *Module) BestTitle(objectID object.ID) string {
	descs := mod.objects.Describe(context.Background(), objectID, &desc.Opts{
		IdentityFilter: id.AllowEveryone,
	})

	var m = map[string]*desc.Desc{}
	for _, desc := range descs {
		t := desc.Data.Type()
		cur, found := m[t]
		if !found {
			m[t] = desc
			continue
		}
		if mod.compareTrust(cur, desc) > 0 {
			m[t] = desc
		}
	}

	if desc, found := m[content.LabelDesc{}.Type()]; found {
		d, _ := desc.Data.(content.LabelDesc)
		return d.Label
	}

	if desc, found := m[keys.KeyDesc{}.Type()]; found {
		d, _ := desc.Data.(keys.KeyDesc)
		return d.String()
	}

	if desc, found := m[fs.FileDesc{}.Type()]; found {
		d, _ := desc.Data.(fs.FileDesc)
		return d.String()
	}

	if desc, found := m[(&media.Audio{}).Type()]; found {
		d, _ := desc.Data.(*media.Audio)
		return d.String()
	}

	if desc, found := m[content.TypeDesc{}.Type()]; found {
		d, _ := desc.Data.(content.TypeDesc)
		return "Untitled " + d.String()
	}

	for _, desc := range descs {
		if s, ok := desc.Data.(fmt.Stringer); ok {
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
