package kos

import (
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
)

const ModuleName = "kos"
const DBPrefix = "kos__"

type Module interface {
	Set(ctx *astral.Context, key string, object astral.Object) error
	Get(ctx *astral.Context, key string) (astral.Object, error)
	Delete(ctx *astral.Context, key string) error
}

func Get[T astral.Object](ctx *astral.Context, mod Module, key string) (t T, err error) {
	obj, err := mod.Get(ctx, key)
	if err != nil {
		return t, err
	}

	t, ok := obj.(T)
	if !ok {
		err = errors.New("typecast failed")
	}

	return t, err
}
