package kos

import (
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
)

const ModuleName = "kos"
const DBPrefix = "kos__"

type Module interface {
	SetObject(ctx *astral.Context, key string, object astral.Object) error
	GetObject(ctx *astral.Context, key string) (astral.Object, error)
	DeleteObject(ctx *astral.Context, key string) error
}

func GetObject[T astral.Object](ctx *astral.Context, mod Module, key string) (t T, err error) {
	obj, err := mod.GetObject(ctx, key)
	if err != nil {
		return t, err
	}

	t, ok := obj.(T)
	if !ok {
		err = errors.New("typecast failed")
	}

	return t, err
}
