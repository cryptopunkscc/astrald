package objects

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/object"
)

func Load[T astral.Object](ctx *astral.Context, mod Module, objectID object.ID, scope *astral.Scope) (o T, err error) {
	if objectID.Size > ReadAllMaxSize {
		return o, ErrObjectTooLarge
	}

	if ctx == nil {
		ctx = astral.NewContext(context.Background())
	}

	r, err := mod.Open(ctx, objectID, &OpenOpts{
		QueryFilter: scope.QueryFilter,
	})

	if err != nil {
		return o, err
	}
	defer r.Close()

	var a astral.Object
	var ok bool

	a, _, err = mod.Blueprints().Read(r, true)
	if err != nil {
		return
	}

	o, ok = a.(T)
	if !ok {
		err = errors.New("typecast failed")
	}

	return
}
