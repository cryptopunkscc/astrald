package shares

import (
	"encoding/json"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"time"
)

type DescriptorCache struct {
	mod *Module
}

func (cache *DescriptorCache) Store(caller id.Identity, target id.Identity, dataID data.ID, list []desc.Data) (err error) {
	var jlist []JSONDescriptor

	for _, item := range list {
		jdesc := JSONDescriptor{
			Type: item.Type(),
		}
		jdesc.Data, err = json.Marshal(item)
		if err != nil {
			cache.mod.log.Error("error serializing JSON descriptor for %s: %s", target, err)
		}
		jlist = append(jlist, jdesc)
	}

	bytes, err := json.Marshal(jlist)
	if err != nil {
		return err
	}

	err = cache.Clear(caller, target, dataID)
	if err != nil {
		return err
	}

	return cache.mod.db.Create(&dbRemoteDesc{
		Caller: caller,
		Target: target,
		DataID: dataID,
		Desc:   string(bytes),
	}).Error
}

func (cache *DescriptorCache) Clear(caller id.Identity, target id.Identity, dataID data.ID) error {
	return cache.mod.db.
		Model(&dbRemoteDesc{}).
		Delete("caller = ? AND target = ? AND data_id = ?", caller, target, dataID).
		Error
}

// Purge deletes all cache entries at least minAge old.
func (cache *DescriptorCache) Purge(minAge time.Duration) error {
	return cache.mod.db.
		Where("created_at < ?", time.Now().Add(-minAge)).
		Delete(&dbRemoteDesc{}).
		Error
}

// Load a cache entry. If maxAge is zero, no age limit will be applied.
func (cache *DescriptorCache) Load(caller id.Identity, target id.Identity, dataID data.ID, maxAge time.Duration) (list []desc.Data, err error) {
	// check cache
	var row dbRemoteDesc
	tx := cache.mod.db.
		Where("caller = ? AND target = ? AND data_id = ?",
			caller,
			target,
			dataID,
		)

	if maxAge > 0 {
		tx = tx.Where("created_at > ?", time.Now().Add(-maxAge))
	}

	err = tx.
		First(&row).Error

	var j []JSONDescriptor
	err = json.Unmarshal([]byte(row.Desc), &j)
	if err != nil {
		return nil, err
	}

	for _, i := range j {
		var d = cache.mod.content.UnmarshalDescriptor(i.Type, i.Data)
		if d == nil {
			continue
		}

		list = append(list, d)
	}

	return list, nil
}
