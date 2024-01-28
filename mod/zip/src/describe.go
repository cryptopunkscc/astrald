package zip

import (
	"context"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/zip"
)

func (mod *Module) Describe(ctx context.Context, dataID data.ID, opts *content.DescribeOpts) (desc []content.Descriptor) {
	desc = append(desc, mod.describeArchive(dataID)...)
	desc = append(desc, mod.describeMember(dataID)...)

	return
}

func (mod *Module) describeArchive(dataID data.ID) []content.Descriptor {
	rows, _ := mod.dbFindByZipID(dataID)
	if len(rows) == 0 {
		return nil
	}

	var desc zip.ArchiveDescriptor

	for _, row := range rows {
		dataID, err := data.Parse(row.FileID)
		if err != nil {
			continue
		}

		desc.Files = append(desc.Files, zip.ArchiveFile{
			DataID: dataID,
			Path:   row.Path,
		})
	}

	return []content.Descriptor{desc}
}

func (mod *Module) describeMember(dataID data.ID) []content.Descriptor {
	rows, _ := mod.dbFindByFileID(dataID)
	if len(rows) == 0 {
		return nil
	}
	var desc zip.MemberDescriptor

	for _, row := range rows {
		zipID, err := data.Parse(row.ZipID)
		if err != nil {
			continue
		}

		desc.Memberships = append(desc.Memberships, zip.Membership{
			ZipID: zipID,
			Path:  row.Path,
		})
	}

	return []content.Descriptor{desc}
}
