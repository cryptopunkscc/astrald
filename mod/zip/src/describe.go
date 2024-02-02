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
	var row dbZip

	var err = mod.db.
		Preload("Contents").
		Where("data_id = ?", dataID).
		First(&row).Error
	if err != nil {
		return nil
	}

	var desc zip.ArchiveDescriptor

	for _, i := range row.Contents {
		desc.Files = append(desc.Files, zip.ArchiveFile{
			DataID: i.FileID,
			Path:   i.Path,
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
		desc.Memberships = append(desc.Memberships, zip.Membership{
			ZipID: row.Zip.DataID,
			Path:  row.Path,
		})
	}

	return []content.Descriptor{desc}
}
