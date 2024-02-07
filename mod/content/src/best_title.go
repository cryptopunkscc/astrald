package content

import (
	"context"
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/mod/media"
)

func (mod *Module) BestTitle(dataID data.ID) string {
	descs := mod.Describe(context.Background(), dataID, &content.DescribeOpts{
		IdentityFilter: id.AllowEveryone,
	})

	var m = map[string]*content.Descriptor{}
	for _, desc := range descs {
		t := desc.Data.DescriptorType()
		cur, found := m[t]
		if !found {
			m[t] = desc
			continue
		}
		if mod.compareTrust(cur, desc) > 0 {
			m[t] = desc
		}
	}

	if desc, found := m[content.LabelDescriptor{}.DescriptorType()]; found {
		d, _ := desc.Data.(content.LabelDescriptor)
		return d.Label
	}

	if desc, found := m[media.Descriptor{}.DescriptorType()]; found {
		d, _ := desc.Data.(media.Descriptor)
		return d.String()
	}

	if desc, found := m[keys.KeyDescriptor{}.DescriptorType()]; found {
		d, _ := desc.Data.(keys.KeyDescriptor)
		return d.String()
	}

	if desc, found := m[fs.FileDescriptor{}.DescriptorType()]; found {
		d, _ := desc.Data.(fs.FileDescriptor)
		return d.String()
	}

	if desc, found := m[content.TypeDescriptor{}.DescriptorType()]; found {
		d, _ := desc.Data.(content.TypeDescriptor)
		return "Untitled " + d.String()
	}

	for _, desc := range descs {
		if s, ok := desc.Data.(fmt.Stringer); ok {
			return s.String()
		}
	}

	return dataID.String()
}

func (mod *Module) compareTrust(a, b *content.Descriptor) int {
	if a.Source.IsEqual(b.Source) {
		return 0
	}

	return -1
}
