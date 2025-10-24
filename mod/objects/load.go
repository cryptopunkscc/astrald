package objects

import (
	"context"
	"fmt"
	"reflect"

	"github.com/cryptopunkscc/astrald/astral"
)

func Load[T astral.Object](ctx *astral.Context, repo Repository, objectID *astral.ObjectID, bp *astral.Blueprints) (o T, err error) {
	if int64(objectID.Size) > MaxObjectSize {
		return o, ErrObjectTooLarge
	}

	if ctx == nil {
		ctx = astral.NewContext(context.Background())
	}

	r, err := repo.Read(ctx, objectID, 0, 0)

	if err != nil {
		return o, err
	}
	defer r.Close()

	var a astral.Object
	var ok bool

	if bp == nil {
		bp = astral.DefaultBlueprints
	}

	a, _, err = bp.ReadCanonical(r)
	if err != nil {
		return
	}

	o, ok = a.(T)
	if !ok {
		err = fmt.Errorf("cannot cast %s into %s", reflect.TypeOf(a), reflect.TypeOf(o))
	}

	return
}
