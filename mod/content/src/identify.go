package content

import (
	"bytes"
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
	"github.com/cryptopunkscc/astrald/sig"
	"github.com/wailsapp/mimetype"
	"time"
)

// Identify returns info about data type.
func (mod *Module) Identify(objectID object.ID) (*content.TypeInfo, error) {
	var err error

	ch, ok := mod.ongoing.Set(objectID.String(), sig.New())
	if ok {
		defer func() {
			mod.ongoing.Delete(objectID.String())
			close(ch)
		}()
	} else {
		<-ch
		if c := mod.getCache(objectID); c != nil {
			return c, nil
		}
		return nil, errors.New("unidentified object")
	}

	if c := mod.getCache(objectID); c != nil {
		return c, nil
	}

	// read first bytes for type identification
	dataReader, err := mod.Objects.Open(context.Background(), objectID, objects.DefaultOpenOpts())
	if err != nil {
		return nil, err
	}

	var firstBytes = make([]byte, identifySize)
	dataReader.Read(firstBytes)
	dataReader.Close()

	var reader = bytes.NewReader(firstBytes)
	var method, dataType string

	// detect type either via adc or mime
	var adcHeader astral.ObjectHeader
	_, err = adcHeader.ReadFrom(reader)
	if err == nil {
		method, dataType = adcMethod, string(adcHeader)
	} else {
		method, dataType = mimetypeMethod, mimetype.Detect(firstBytes).String()
	}

	var indexedAt = time.Now()
	var tx = mod.db.Create(&dbDataType{
		DataID:       objectID,
		IdentifiedAt: indexedAt,
		Method:       method,
		Type:         dataType,
	})
	if tx.Error != nil {
		return nil, tx.Error
	}

	mod.log.Logv(1, "%v identified as %s via %s", objectID, dataType, method)

	info := &content.TypeInfo{
		ObjectID:     objectID,
		IdentifiedAt: indexedAt,
		Method:       method,
		Type:         dataType,
	}

	mod.events.Emit(content.EventObjectIdentified{TypeInfo: info})

	return info, nil
}

func (mod *Module) getCache(objectID object.ID) *content.TypeInfo {
	var row dbDataType

	err := mod.db.Where("data_id = ?", objectID).First(&row).Error
	if err != nil {
		return nil
	}

	return &content.TypeInfo{
		ObjectID:     row.DataID,
		IdentifiedAt: row.IdentifiedAt,
		Method:       row.Method,
		Type:         row.Type,
	}
}

func (mod *Module) identifyFS() {
	if mod.FS == nil {
		return
	}

	for _, file := range mod.FS.Find(nil) {
		mod.Identify(file.ObjectID)
	}
}
