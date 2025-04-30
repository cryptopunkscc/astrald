package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
	"sync"
)

type RootRepository struct {
	mod *Module
}

var _ objects.Repository = &RootRepository{}

func NewRootRepository(mod *Module) *RootRepository {
	return &RootRepository{mod: mod}
}

func (repo RootRepository) Label() string {
	return "Default Repository"
}

func (repo RootRepository) Create(ctx *astral.Context, opts *objects.CreateOpts) (objects.Writer, error) {
	return repo.Default().Create(ctx, opts)
}

func (repo RootRepository) Contains(ctx *astral.Context, objectID *object.ID) (bool, error) {
	for _, repo := range repo.mod.repos.Clone() {
		if has, err := repo.Contains(ctx, objectID); err == nil && has {
			return true, nil
		}
	}

	return false, nil
}

func (repo RootRepository) Scan(ctx *astral.Context, follow bool) (<-chan *object.ID, error) {
	ch := make(chan *object.ID)

	go func() {
		var wg sync.WaitGroup

		for _, repo := range repo.mod.repos.Clone() {
			wg.Add(1)
			go func() {
				defer wg.Done()

				sub, err := repo.Scan(ctx, follow)
				if err != nil {
					return
				}

				var id *object.ID

				// copy all scanned ids
				for {
					// read
					select {
					case <-ctx.Done():
						return
					case id = <-sub:
					}
					// write
					select {
					case <-ctx.Done():
						return
					case ch <- id:
					}
				}
			}()
		}

		wg.Wait()
		close(ch)
	}()

	return ch, nil
}

func (repo RootRepository) Delete(ctx *astral.Context, objectID *object.ID) error {
	return repo.Default().Delete(ctx, objectID)
}

func (repo RootRepository) Read(ctx *astral.Context, objectID *object.ID, offset int64, limit int64) (objects.Reader, error) {
	// try memory cache first
	if mem, err := repo.mod.GetRepository("mem0"); err == nil {
		if r, err := mem.Read(ctx, objectID, offset, limit); err == nil {
			return r, nil
		}
	}

	// then default storage
	if mem, err := repo.mod.GetRepository("default"); err == nil {
		if r, err := mem.Read(ctx, objectID, offset, limit); err == nil {
			return r, nil
		}
	}

	// then all other repos
	for id, repo := range repo.mod.repos.Clone() {
		if id == "mem0" || id == "default" {
			continue
		}

		if r, err := repo.Read(ctx, objectID, offset, limit); err == nil {
			return r, nil
		}
	}

	r, err := repo.readNetwork(ctx, objectID, offset, limit)
	if err == nil {
		return r, nil
	}

	return nil, objects.ErrNotFound
}

func (repo *RootRepository) readNetwork(ctx *astral.Context, objectID *object.ID, offset int64, limit int64) (objects.Reader, error) {
	if !ctx.Zone().Is(astral.ZoneNetwork) {
		return nil, astral.ErrZoneExcluded
	}

	providers := repo.mod.Find(ctx, objectID)

	var conns = make(chan objects.Reader, 1)
	var wg sync.WaitGroup

	ctx, done := ctx.WithCancel()
	defer done()

	for _, providerID := range providers {
		providerID := providerID

		wg.Add(1)
		go func() {
			defer wg.Done()

			c := NewConsumer(repo.mod, repo.mod.node.Identity(), providerID)

			r, err := c.Read(ctx, objectID, offset, limit)
			if err != nil {
				return
			}

			select {
			case conns <- r:
				done()
			default:
				r.Close()
			}
		}()
	}

	go func() {
		wg.Wait()
		close(conns)
	}()

	r, ok := <-conns
	if !ok {
		return nil, objects.ErrNotFound
	}

	return r, nil
}

func (repo RootRepository) Free(ctx *astral.Context) (int64, error) {
	return repo.Default().Free(ctx)
}

func (repo RootRepository) Default() (r objects.Repository) {
	r, err := repo.mod.GetRepository("default")
	if err == nil {
		return
	}

	r, err = repo.mod.GetRepository("mem0")
	if err != nil {
		panic(err)
	}

	return
}
