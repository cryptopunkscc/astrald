package objects

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/object"
)

func Load[T astral.Object](ctx context.Context, mod Module, objectID object.ID, scope *astral.Scope) (o T, err error) {
	if objectID.Size > ReadAllMaxSize {
		return o, ErrObjectTooLarge
	}

	r, err := mod.Open(ctx, objectID, &OpenOpts{
		Zone:        scope.Zone,
		QueryFilter: scope.QueryFilter,
	})

	if err != nil {
		return o, err
	}
	defer r.Close()

	var a astral.Object
	var ok bool

	a, err = mod.ReadObject(r)
	if err != nil {
		return
	}

	o, ok = a.(T)
	if !ok {
		err = errors.New("typecast failed")
	}

	return
}
