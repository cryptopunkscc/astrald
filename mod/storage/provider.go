package storage

import "github.com/cryptopunkscc/astrald/auth/id"

func (mod *Module) AddProvider(identity id.Identity) error {
	return mod.db.Create(&dbProvider{Identity: identity.String()}).Error
}

func (mod *Module) RemoveProvider(identity id.Identity) error {
	return mod.db.Delete(&dbProvider{Identity: identity.String()}).Error
}

func (mod *Module) IsProvider(identity id.Identity) bool {
	var c int64
	mod.db.Model(&dbProvider{}).Where("identity = ?", identity.String()).Count(&c)
	return c > 0
}

func (mod *Module) AllProviders() ([]id.Identity, error) {
	var rows []dbProvider
	if err := mod.db.Find(&rows).Error; err != nil {
		return nil, err
	}

	var list = make([]id.Identity, 0, len(rows))
	for _, row := range rows {
		i, err := id.ParsePublicKeyHex(row.Identity)
		if err != nil {
			continue
		}
		list = append(list, i)
	}
	return list, nil
}
