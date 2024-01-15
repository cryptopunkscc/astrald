package zip

import (
	"context"
	_data "github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/data"
	"github.com/cryptopunkscc/astrald/mod/zip"
)

func (mod *Module) DescribeData(ctx context.Context, dataID _data.ID, opts *data.DescribeOpts) (desc []data.Descriptor) {
	desc = append(desc, mod.describeArchive(dataID)...)
	desc = append(desc, mod.describeMember(dataID)...)

	return
}

func (mod *Module) describeArchive(dataID _data.ID) []data.Descriptor {
	rows, _ := mod.dbFindByZipID(dataID)
	if len(rows) == 0 {
		return nil
	}

	var desc zip.ArchiveDescriptor

	for _, row := range rows {
		dataID, err := _data.Parse(row.FileID)
		if err != nil {
			continue
		}

		desc.Files = append(desc.Files, zip.ArchiveFile{
			DataID: dataID,
			Path:   row.Path,
		})
	}

	return []data.Descriptor{{
		Type: zip.ArchiveDescriptorType,
		Data: desc,
	}}
}

func (mod *Module) describeMember(dataID _data.ID) []data.Descriptor {
	rows, _ := mod.dbFindByFileID(dataID)
	if len(rows) == 0 {
		return nil
	}
	var desc zip.MemberDescriptor

	for _, row := range rows {
		zipID, err := _data.Parse(row.ZipID)
		if err != nil {
			continue
		}

		desc.Memberships = append(desc.Memberships, zip.Membership{
			ZipID: zipID,
			Path:  row.Path,
		})
	}

	return []data.Descriptor{{
		Type: zip.MemberDescriptorType,
		Data: desc,
	}}
}
