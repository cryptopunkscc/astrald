package media

import (
	"fmt"
	"strings"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/dir"
	media "github.com/cryptopunkscc/astrald/mod/media"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

var _ media.Module = &Module{}
var _ objects.Searcher = &Module{}
var _ objects.Describer = &Module{}

type Deps struct {
	Dir dir.Module
}

type Module struct {
	Deps
	config Config
	node   astral.Node
	log    *log.Logger
}

func (mod *Module) Run(*astral.Context) error {
	return nil
}

func (mod *Module) String() string {
	return media.ModuleName
}

func (mod *Module) resolveTarget() (*astral.Identity, error) {
	target := strings.TrimSpace(mod.config.App)
	if target == "" {
		return nil, fmt.Errorf("%s app target is not configured", media.ModuleName)
	}

	targetID, err := mod.Dir.ResolveIdentity(target)
	if err != nil {
		return nil, fmt.Errorf("resolve %s app %q: %w", media.ModuleName, target, err)
	}

	return targetID, nil
}
