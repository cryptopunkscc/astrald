package content

import (
	"bytes"
	"github.com/cryptopunkscc/astrald/lib/adc"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
	"github.com/wailsapp/mimetype"
	"time"
)

// Identify returns info about data type.
func (mod *Module) Identify(objectID object.ID) (*content.TypeInfo, error) {
	var err error
	var row dbDataType

	// check if data is already indexed
	err = mod.db.Where("data_id = ?", objectID).First(&row).Error
	if err == nil {
		return &content.TypeInfo{
			ObjectID:     objectID,
			IdentifiedAt: row.IdentifiedAt,
			Method:       row.Method,
			Type:         row.Type,
		}, nil
	}

	// read first bytes for type identification
	dataReader, err := mod.objects.Open(objectID, &objects.OpenOpts{Virtual: true, Network: true})
	if err != nil {
		return nil, err
	}

	var firstBytes = make([]byte, identifySize)
	dataReader.Read(firstBytes)
	dataReader.Close()

	var reader = bytes.NewReader(firstBytes)
	var method, dataType string

	// detect type either via adc or mime
	adcHeader, err := adc.ReadHeader(reader)
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

func (mod *Module) identifyFS() {
	if mod.fs == nil {
		return
	}

	for _, file := range mod.fs.Find(nil) {
		mod.Identify(file.ObjectID)
	}
}
