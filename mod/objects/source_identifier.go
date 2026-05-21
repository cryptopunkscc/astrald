package objects

import "github.com/cryptopunkscc/astrald/astral"

// SourceIdentifier marks discovery adapters by their source identity;
// "external" may be the better conceptual name.
type SourceIdentifier interface {
	SourceIdentity() *astral.Identity
}

func SourceIdentity(v any) (*astral.Identity, bool, error) {
	source, ok := v.(SourceIdentifier)
	if !ok {
		return nil, false, nil
	}

	if source == nil {
		return nil, true, ErrNilSourceIdentifier
	}

	id := source.SourceIdentity()
	if id == nil || id.IsZero() {
		return nil, true, ErrInvalidSourceIdentity
	}

	return id, true, nil
}
