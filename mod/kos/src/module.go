package kos

import (
	"bytes"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/kos"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/resources"
)

type Module struct {
	Deps
	config Config
	node   astral.Node
	log    *log.Logger
	db     *DB
	ops    shell.Scope
	assets resources.Resources
}

var _ kos.Module = &Module{}

func (mod *Module) Run(ctx *astral.Context) error {
	<-ctx.Done()
	return nil
}

func (mod *Module) SetObject(ctx *astral.Context, key string, object astral.Object) (err error) {
	var buf = &bytes.Buffer{}

	_, err = object.WriteTo(buf)
	if err != nil {
		return err
	}

	return mod.db.Set(ctx.Identity(), key, object.ObjectType(), buf.Bytes())
}

func (mod *Module) GetObject(ctx *astral.Context, key string) (astral.Object, error) {
	typ, payload, err := mod.db.Get(ctx.Identity(), key)
	if err != nil {
		return nil, err
	}

	object := mod.Objects.Blueprints().Make(typ)
	if object == nil {
		object = &astral.RawObject{}
	}

	_, err = object.ReadFrom(bytes.NewReader(payload))

	return object, err
}

func (mod *Module) DeleteObject(ctx *astral.Context, key string) error {
	return mod.db.Delete(ctx.Identity(), key)
}

func (mod *Module) Scope() *shell.Scope {
	return &mod.ops
}

func (mod *Module) String() string {
	return kos.ModuleName
}
