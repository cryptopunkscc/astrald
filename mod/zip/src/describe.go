package zip

import (
	"context"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"github.com/cryptopunkscc/astrald/mod/zip"
)

func (mod *Module) Describe(ctx context.Context, dataID data.ID, opts *desc.Opts) (desc []*desc.Desc) {
	desc = append(desc, mod.describeArchive(dataID)...)
	desc = append(desc, mod.describeMember(dataID)...)

	return
}

func (mod *Module) describeArchive(dataID data.ID) []*desc.Desc {
	var row dbZip

	var err = mod.db.
		Preload("Contents").
		Where("data_id = ?", dataID).
		First(&row).Error
	if err != nil {
		return nil
	}

	var ad zip.ArchiveDesc

	for _, i := range row.Contents {
		ad.Files = append(ad.Files, zip.ArchiveFile{
			DataID: i.FileID,
			Path:   i.Path,
		})
	}

	return []*desc.Desc{{
		Source: mod.node.Identity(),
		Data:   ad,
	}}
}

func (mod *Module) describeMember(dataID data.ID) []*desc.Desc {
	rows, _ := mod.dbFindByFileID(dataID)
	if len(rows) == 0 {
		return nil
	}
	var ad zip.MemberDesc

	for _, row := range rows {
		ad.Memberships = append(ad.Memberships, zip.Membership{
			ZipID: row.Zip.DataID,
			Path:  row.Path,
		})
	}

	return []*desc.Desc{{
		Source: mod.node.Identity(),
		Data:   ad,
	}}
}
