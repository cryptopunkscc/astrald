package dir

import (
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/resources"
	"github.com/cryptopunkscc/astrald/sig"
	"gorm.io/gorm"
	"os"
)

var _ dir.Module = &Module{}

const ZeroIdentity = "<anyone>"

type Module struct {
	config Config
	node   astral.Node
	log    *log.Logger
	assets resources.Resources
	db     *gorm.DB

	resolvers  sig.Set[dir.Resolver]
	describers sig.Set[dir.Describer]
}

func (mod *Module) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (mod *Module) AddResolver(resolver dir.Resolver) error {
	return mod.resolvers.Add(resolver)
}

func (mod *Module) Resolve(s string) (identity id.Identity, err error) {
	if s == "" || s == "anyone" {
		return id.Identity{}, nil
	}

	if s == "localnode" {
		return mod.node.Identity(), nil
	}

	if identity, err := id.ParsePublicKeyHex(s); err == nil {
		return identity, nil
	}

	err = mod.db.
		Model(&dbAlias{}).
		Where("alias = ?", s).
		Select("identity").
		First(&identity).
		Error
	if err == nil {
		return
	}

	for _, r := range mod.resolvers.Clone() {
		if i, err := r.Resolve(s); err == nil {
			return i, nil
		}
	}

	return id.Identity{}, fmt.Errorf("unknown identity: %s", s)
}

func (mod *Module) DisplayName(identity id.Identity) string {
	if identity.IsZero() {
		return ZeroIdentity
	}

	a, err := mod.GetAlias(identity)
	if err == nil {
		return a
	}

	for _, r := range mod.resolvers.Clone() {
		if s := r.DisplayName(identity); len(s) > 0 {
			return s
		}
	}

	return identity.Fingerprint()
}

func (mod *Module) Describe(ctx context.Context, identity id.Identity, opts *desc.Opts) []*desc.Desc {
	var list []desc.Describer[id.Identity]

	for _, d := range mod.describers.Clone() {
		list = append(list, d)
	}

	return desc.Collect(ctx, identity, opts, list...)
}

func (mod *Module) AddDescriber(describer dir.Describer) error {
	return mod.describers.Add(describer)
}

func (mod *Module) RemoveDescriber(describer dir.Describer) error {
	return mod.describers.Remove(describer)
}

func (mod *Module) setDefaultAlias() error {
	alias, err := mod.GetAlias(mod.node.Identity())
	if (err != nil) && (!errors.Is(err, gorm.ErrRecordNotFound)) {
		return err
	}
	if alias != "" {
		return nil
	}

	alias = "localnode"

	hostname, err := os.Hostname()
	if err == nil {
		if hostname != "" && hostname != "localhost" {
			alias = hostname
		}
	}

	return mod.SetAlias(mod.node.Identity(), alias)
}
