package media

import (
	"context"
	"github.com/acuteaura-forks/go-matroska/ebml"
	"github.com/acuteaura-forks/go-matroska/matroska"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/media"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"io"
	"strings"
	"time"
)

var _ Indexer = &MatroskaIndexer{}

type MatroskaIndexer struct {
	*Module
}

func NewMatroskaIndexer(mod *Module) *MatroskaIndexer {
	return &MatroskaIndexer{Module: mod}
}

func (mod *MatroskaIndexer) Describe(ctx context.Context, dataID data.ID, opts *desc.Opts) []*desc.Desc {
	var row dbVideo
	var err = mod.db.Where("data_id = ?", dataID).First(&row).Error
	var info *media.Video
	if err == nil {
		info = &media.Video{
			Format:   row.Format,
			Title:    row.Title,
			Duration: row.Duration,
		}
	} else {
		info, err = mod.index(dataID, &storage.OpenOpts{
			Virtual: true,
		})
		if err != nil {
			mod.log.Errorv(2, "error indexing %v: %v", dataID, err)
		}
	}
	if info == nil {
		return nil
	}

	return []*desc.Desc{{
		Source: mod.node.Identity(),
		Data:   info,
	}}
}

func (mod *MatroskaIndexer) Find(ctx context.Context, query string, opts *content.FindOpts) (matches []content.Match, err error) {
	var rows []*dbVideo

	query = "%" + strings.ToLower(query) + "%"

	err = mod.db.
		Where("LOWER(title) LIKE ?", query).
		Find(&rows).
		Error
	if err != nil {
		mod.log.Error("db error: %v", err)
		return
	}

	for _, row := range rows {
		matches = append(matches, content.Match{
			DataID: row.DataID,
			Score:  100,
			Exp:    "video matches query",
		})
	}

	return
}

func (mod *MatroskaIndexer) index(dataID data.ID, opts *storage.OpenOpts) (*media.Video, error) {
	r, err := mod.storage.Open(dataID, opts)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	info, err := mod.scan(r)
	if err != nil {
		return nil, err
	}

	err = mod.db.Create(&dbVideo{
		DataID:   dataID,
		Format:   info.Format,
		Title:    info.Title,
		Duration: info.Duration,
	}).Error

	return info, err
}

func (mod *MatroskaIndexer) scan(r io.Reader) (*media.Video, error) {
	file, err := DecodeMKV(r)
	if err != nil {
		return nil, err
	}

	var info = &media.Video{
		Format: "mkv",
	}

	if file.Segment != nil || len(file.Segment.Info) > 0 {
		i := file.Segment.Info[0]
		info.Title = i.Title
		info.Duration = time.Duration(i.Duration) * time.Millisecond
	}

	return info, err
}

func DecodeMKV(r io.Reader) (*matroska.File, error) {
	dec := ebml.NewReader(r, &ebml.DecodeOptions{
		SkipDamaged: true,
	})

	v := new(matroska.File)
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}
	return v, nil
}
