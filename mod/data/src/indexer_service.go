package data

import (
	"bytes"
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/data"
	_data "github.com/cryptopunkscc/astrald/mod/data"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/wailsapp/mimetype"
	"path/filepath"
	"time"
)

type IndexService struct {
	*Module
}

func (srv *IndexService) Run(ctx context.Context) error {
	var t time.Time

	if c, err := srv.mostRecentContainer(); err == nil {
		t = c.IndexedAt
	}

	go srv.handleEvents(ctx)

	for _, d := range srv.storage.Data().IndexSince(t) {
		srv.index(d)
	}

	<-ctx.Done()
	return nil
}

func (srv *IndexService) handleEvents(ctx context.Context) {
	for event := range srv.node.Events().Subscribe(ctx) {
		switch event := event.(type) {
		case storage.EventDataAdded:
			srv.index(storage.IndexEntry(event))
		}
	}
}

func (srv *IndexService) index(info storage.IndexEntry) error {
	// skip already indexed data
	if _, err := srv.findByDataID(info.ID); err == nil {
		return errors.New("already indexed")
	}

	dataReader, err := srv.storage.Data().Read(info.ID, nil)
	if err != nil {
		return err
	}

	var firstBytes = make([]byte, 512)
	dataReader.Read(firstBytes)
	dataReader.Close()

	var (
		reader     = bytes.NewReader(firstBytes)
		adc0Header _data.ADC0Header
		dataType   string
		header     string
	)

	if err := cslq.Decode(reader, "v", &adc0Header); err == nil {
		dataType = string(adc0Header)
		header = "adc0"
	} else {
		dataType = mimetype.Detect(firstBytes).String()
	}

	srv.db.Create(&dbDataType{
		DataID:    info.ID.String(),
		IndexedAt: info.IndexedAt,
		Header:    header,
		Type:      dataType,
	})

	if header != "" {
		srv.log.Logv(2, "identified %v as %s (%s)", info.ID, dataType, header)
	} else {
		srv.log.Logv(2, "identified %v as %s", info.ID, dataType)
	}

	if name := srv.getFileName(info.ID); name != "" {
		srv.SetLabel(info.ID, name)
	}

	srv.events.Emit(_data.EventDataIndexed{
		DataID:    info.ID,
		IndexedAt: info.IndexedAt,
		Header:    header,
		Type:      dataType,
	})

	return nil
}

func (srv *IndexService) getFileName(dataID data.ID) string {
	if srv.fs == nil {
		return ""
	}

	for _, path := range srv.fs.Find(dataID) {
		return filepath.Base(path)
	}

	return ""
}
