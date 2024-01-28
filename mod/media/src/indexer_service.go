package media

import (
	"context"
	"errors"
	"github.com/acuteaura-forks/go-matroska/ebml"
	"github.com/acuteaura-forks/go-matroska/matroska"
	"github.com/bogem/id3v2"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/media"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"io"
	"slices"
	"time"
)

type IndexerService struct {
	*Module
}

var networkAutoIndexWhitelist = []string{
	"audio/mpeg",
}

func (srv *IndexerService) Run(ctx context.Context) error {
	for event := range srv.content.Scan(ctx, nil) {
		found, err := srv.index.Contains(media.IndexNameAll, event.DataID)
		if found {
			continue
		}

		_, err = srv.autoIndex(event.DataID, event.Type)
		if err != nil && !errors.Is(err, storage.ErrNotFound) {
			srv.log.Error("error indexing %v: %v", event.DataID, err)
		}
	}

	<-ctx.Done()

	return nil
}

func (srv *IndexerService) autoIndex(dataID data.ID, dataType string) (*media.Info, error) {
	var enableNetwork bool

	if slices.Contains(networkAutoIndexWhitelist, dataType) {
		enableNetwork = true
	}

	return srv.indexAs(dataID, dataType, enableNetwork)
}

func (srv *IndexerService) indexAs(dataID data.ID, dataType string, enableNetwork bool) (*media.Info, error) {
	r, err := srv.storage.Read(
		dataID,
		&storage.ReadOpts{
			Virtual: true,
			Network: enableNetwork,
		},
	)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var start = time.Now()
	info, err := srv.scanAs(r, dataType)
	var d = time.Since(start)
	if err != nil {
		return nil, err
	}
	if info == nil {
		return nil, nil
	}

	srv.log.Infov(2, "scanned %v in %v", dataID, d)

	err = srv.db.Create(&dbMediaInfo{
		DataID: dataID,
		Type:   info.Type,
		Artist: info.Artist,
		Title:  info.Title,
		Album:  info.Album,
		Genre:  info.Genre,
	}).Error
	if err != nil {
		return info, err
	}

	if srv.content.GetLabel(dataID) == "" {
		srv.content.SetLabel(dataID, info.Title)
	}

	return info, srv.index.AddToSet(media.IndexNameAll, dataID)
}

func (srv *IndexerService) indexData(dataID data.ID) (*media.Info, error) {
	info, err := srv.content.Identify(dataID)
	if err != nil {
		return nil, err
	}

	return srv.indexAs(dataID, info.Type, true)
}

func (srv *IndexerService) scan(dataID data.ID) (*media.Info, error) {
	info, err := srv.content.Identify(dataID)
	if err != nil {
		return nil, err
	}

	r, err := srv.storage.Read(dataID, &storage.ReadOpts{Virtual: true})
	if err != nil {
		return nil, err
	}
	defer r.Close()

	return srv.scanAs(r, info.Type)
}

func (srv *IndexerService) scanAs(r io.Reader, dataType string) (info *media.Info, err error) {
	switch dataType {
	case "video/x-matroska":
		info, err = srv.scanMatroska(r)

	case "audio/mpeg":
		info, err = srv.scanID3v2(r)
	}

	return info, err
}

func (srv *IndexerService) scanID3v2(r io.Reader) (*media.Info, error) {
	tag, err := id3v2.ParseReader(r, id3v2.Options{Parse: true})
	if err != nil {
		return nil, err
	}

	return &media.Info{
		Type:   "audio",
		Title:  tag.Title(),
		Artist: tag.Artist(),
		Album:  tag.Album(),
		Genre:  tag.Genre(),
	}, err
}

func (srv *IndexerService) scanMatroska(r io.Reader) (*media.Info, error) {
	file, err := DecodeMKV(r)
	if err != nil {
		return nil, err
	}

	var info = &media.Info{
		Type: "video",
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
