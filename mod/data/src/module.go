package data

import (
	"bytes"
	"context"
	"github.com/cryptopunkscc/astrald/cslq"
	_data "github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/data"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/mod/index"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/sig"
	"github.com/wailsapp/mimetype"
	"gorm.io/gorm"
	"time"
)

var _ data.Module = &Module{}

type Module struct {
	node   node.Node
	config Config
	log    *log.Logger
	events events.Queue
	db     *gorm.DB

	describers sig.Set[data.Describer]

	storage storage.Module
	fs      fs.Module
	index   index.Module

	ready chan struct{}
}

func (mod *Module) Run(ctx context.Context) error {
	<-ctx.Done()

	return nil
}

func (mod *Module) StoreADC0(t string, alloc int) (storage.DataWriter, error) {
	w, err := mod.storage.Data().Store(
		&storage.StoreOpts{
			Alloc: alloc + len(t) + 5,
		},
	)
	if err != nil {
		return nil, err
	}

	err = cslq.Encode(w, "v", data.ADC0Header(t))
	if err != nil {
		return nil, err
	}

	return w, nil
}

func (mod *Module) Events() *events.Queue {
	return &mod.events
}

func (mod *Module) Index(dataID _data.ID) error {
	// check if data is already indexed
	if _, err := mod.dbDataTypeFindByDataID(dataID.String()); err == nil {
		return data.ErrAlreadyIndexed
	}

	dataReader, err := mod.storage.Data().Read(dataID, nil)
	if err != nil {
		return err
	}

	var firstBytes = make([]byte, 512)
	dataReader.Read(firstBytes)
	dataReader.Close()

	var (
		reader     = bytes.NewReader(firstBytes)
		adc0Header data.ADC0Header
		dataType   string
		header     = "mimetype"
	)

	// detect type either via adc0 or mime
	if err := cslq.Decode(reader, "v", &adc0Header); err == nil {
		dataType = string(adc0Header)
		header = "adc0"
	} else {
		dataType = mimetype.Detect(firstBytes).String()
	}

	var indexedAt = time.Now()

	var tx = mod.db.Create(&dbDataType{
		DataID:    dataID.String(),
		IndexedAt: indexedAt,
		Header:    header,
		Type:      dataType,
	})
	if tx.Error != nil {
		return tx.Error
	}

	if header != "" {
		mod.log.Logv(1, "%v indexed as %s (%s)", dataID, dataType, header)
	} else {
		mod.log.Logv(1, "%v indexed as %s", dataID, dataType)
	}

	if err := mod.index.AddToSet(data.IdentifiedDataIndexName, dataID); err != nil {
		mod.log.Error("error adding to set: %v", err)
	}

	mod.events.Emit(data.EventDataIdentified{
		DataID:    dataID,
		Header:    header,
		Type:      dataType,
		IndexedAt: indexedAt,
	})

	return nil
}

func (mod *Module) Ready(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()

	case <-mod.ready:
		return nil
	}
}

func (mod *Module) setReady() {
	close(mod.ready)
}
