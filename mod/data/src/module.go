package data

import (
	"bytes"
	"context"
	_data "github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/lib/adc"
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

const identifySize = 4096
const adcMethod = "adc"
const mimetypeMethod = "mimetype"

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

func (mod *Module) Events() *events.Queue {
	return &mod.events
}

// SubscribeType returns a channel that will be populated with all data entries since the provided timestamp and
// subscribed to any new items until context is done. If typ is empty, all data entries will be passed regardless
// of the type.
func (mod *Module) SubscribeType(ctx context.Context, typ string, since time.Time) <-chan data.TypeInfo {
	if since.After(time.Now()) {
		return nil
	}

	var ch = make(chan data.TypeInfo)
	var subscription = mod.events.Subscribe(ctx)

	go func() {
		defer close(ch)

		// catch up with existing entries
		list, err := mod.FindByType(typ, since)
		if err != nil {
			return
		}
		for _, item := range list {
			select {
			case ch <- item:
			case <-ctx.Done():
				return
			}
		}

		// subscribe to new items
		for event := range subscription {
			e, ok := event.(data.EventDataIdentified)
			if !ok {
				continue
			}
			if typ != "" && e.Type != typ {
				continue
			}
			ch <- data.TypeInfo(e)
		}
	}()

	return ch
}

// FindByType returns all data items indexed since time ts. If t is not empty, only items of type t will
// be returned.
func (mod *Module) FindByType(t string, ts time.Time) ([]data.TypeInfo, error) {
	var list []data.TypeInfo
	var rows []*dbDataType

	// filter by type if provided
	var q = mod.db
	if t != "" {
		q = q.Where("type = ?", t)
	}

	// filter by time if provided
	if !ts.IsZero() {
		q = q.Where("indexed_at > ?", ts)
	}

	// fetch rows
	var tx = q.Order("indexed_at").Find(&rows)
	if tx.Error != nil {
		return nil, tx.Error
	}

	for _, row := range rows {
		dataID, err := _data.Parse(row.DataID)
		if err != nil {
			continue
		}

		list = append(list, data.TypeInfo{
			DataID:    dataID,
			IndexedAt: row.IndexedAt,
			Header:    row.Header,
			Type:      row.Type,
		})
	}

	return list, nil
}

// Identify identifies and indexes data type. If data is already indexed, it returns
// ErrAlreadyIndexed.
func (mod *Module) Identify(dataID _data.ID) error {
	// check if data is already indexed
	_, err := mod.dbDataTypeFindByDataID(dataID.String())
	if err == nil {
		return data.ErrAlreadyIndexed
	}

	// read first bytes for type identification
	dataReader, err := mod.storage.Read(dataID, nil)
	if err != nil {
		return err
	}

	var firstBytes = make([]byte, identifySize)
	dataReader.Read(firstBytes)
	dataReader.Close()

	var reader = bytes.NewReader(firstBytes)
	var header, dataType string

	// detect type either via adc or mime
	adcHeader, err := adc.ReadHeader(reader)
	if err == nil {
		header, dataType = adcMethod, string(adcHeader)
	} else {
		header, dataType = mimetypeMethod, mimetype.Detect(firstBytes).String()
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
