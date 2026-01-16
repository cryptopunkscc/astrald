package fs

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/sig"
)

var _ fs.Module = &Module{}

type Module struct {
	Deps
	config  Config
	node    astral.Node
	assets  assets.Assets
	log     *log.Logger
	db      *DB
	ctx     *astral.Context
	indexer *Indexer

	repos sig.Map[string, objects.Repository]
	ops   shell.Scope
}

func (mod *Module) Run(ctx *astral.Context) error {
	mod.ctx = ctx

	go mod.indexer.init(ctx)

	<-ctx.Done()
	return nil
}

func (mod *Module) Scope() *shell.Scope {
	return &mod.ops
}

func (mod *Module) String() string {
	return fs.ModuleName
}

func resolveFileID(path string) (*astral.ObjectID, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileID, err := astral.Resolve(file)
	if err != nil {
		return nil, err
	}

	return fileID, nil
}

func pathUnderRoot(path string, root string) bool {
	root = filepath.Clean(root)
	path = filepath.Clean(path)

	if path == root {
		return true
	}

	return strings.HasPrefix(path, root+string(filepath.Separator))

}
