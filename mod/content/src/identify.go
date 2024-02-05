package content

import (
	"bytes"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/lib/adc"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/wailsapp/mimetype"
	"time"
)

// Identify returns info about data type.
func (mod *Module) Identify(dataID data.ID) (*content.TypeInfo, error) {
	var err error
	var row dbDataType

	// check if data is already indexed
	err = mod.db.Where("data_id = ?", dataID).First(&row).Error
	if err == nil {
		return &content.TypeInfo{
			DataID:       dataID,
			IdentifiedAt: row.IdentifiedAt,
			Method:       row.Method,
			Type:         row.Type,
		}, nil
	}

	// read first bytes for type identification
	dataReader, err := mod.storage.Open(dataID, &storage.OpenOpts{Virtual: true, Network: true})
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
		DataID:       dataID,
		IdentifiedAt: indexedAt,
		Method:       method,
		Type:         dataType,
	})
	if tx.Error != nil {
		return nil, tx.Error
	}

	if method != "" {
		mod.log.Logv(1, "%v identified as %s (%s)", dataID, dataType, method)
	} else {
		mod.log.Logv(1, "%v identified as %s", dataID, dataType)
	}

	if err := mod.identified.Add(dataID); err != nil {
		mod.log.Error("error adding to set: %v", err)
	}

	info := &content.TypeInfo{
		DataID:       dataID,
		IdentifiedAt: indexedAt,
		Method:       method,
		Type:         dataType,
	}

	mod.events.Emit(content.EventDataIdentified{TypeInfo: info})

	return info, nil
}

// IdentifySet identifies all data objects in a set
func (mod *Module) IdentifySet(name string) ([]*content.TypeInfo, error) {
	var list []*content.TypeInfo

	set, err := mod.sets.Open(name)
	if err != nil {
		return nil, err
	}

	entries, err := set.Scan(nil)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		info, err := mod.Identify(entry.DataID)
		if err != nil {
			continue
		}
		list = append(list, info)
	}

	return list, nil
}
