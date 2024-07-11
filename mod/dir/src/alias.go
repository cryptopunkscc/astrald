package dir

import "github.com/cryptopunkscc/astrald/auth/id"

// SetAlias sets the alias for the identity. Set an empty alias to unset.
func (mod *Module) SetAlias(identity id.Identity, alias string) error {
	if alias == "" {
		return mod.db.Delete(&dbAlias{}, "identity = ?", identity).Error
	}

	return mod.db.Save(&dbAlias{
		Identity: identity,
		Alias:    alias,
	}).Error
}

// GetAlias returns the alias for the identity. Returns an empty string if no alias is set.
func (mod *Module) GetAlias(identity id.Identity) (string, error) {
	var row dbAlias
	if err := mod.db.First(&row, "identity = ?", identity).Error; err != nil {
		return "", err
	}

	return row.Alias, nil
}
