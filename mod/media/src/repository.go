package media

import (
	"bytes"
	"io"
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/objects/mem"
	"github.com/cryptopunkscc/astrald/sig"
	"github.com/dhowden/tag"
)

type Repository struct {
	objects.NilRepository
	mod       *Module
	scanQueue *sig.Queue[*astral.ObjectID]
	mu        sync.Mutex
}

func NewRepository(mod *Module) *Repository {
	return &Repository{
		mod:       mod,
		scanQueue: &sig.Queue[*astral.ObjectID]{},
	}
}

var _ objects.Repository = &Repository{}

func (repo *Repository) Label() string {
	return "Media covers"
}

func (repo *Repository) Read(ctx *astral.Context, objectID *astral.ObjectID, offset int64, limit int64) (io.ReadCloser, error) {
	containerID, err := repo.mod.db.FindAudioContainerID(objectID)
	if err != nil {
		return nil, err
	}
	if containerID.IsZero() {
		return nil, objects.ErrNotFound
	}

	r, err := repo.mod.Objects.Root().Read(ctx, containerID, 0, 0)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	audioTag, err := tag.ReadFrom(objects.NewReadSeeker(ctx, containerID, repo.mod.Objects.Root(), r))
	if err != nil {
		return nil, err
	}

	if audioTag.Picture() == nil {
		return nil, objects.ErrNotFound
	}

	pic := audioTag.Picture().Data

	actualID, _ := astral.Resolve(bytes.NewReader(pic))
	if !actualID.IsEqual(objectID) {
		return nil, objects.ErrNotFound
	}

	if limit == 0 {
		limit = int64(len(pic))
	}
	end := min(offset+limit, int64(len(pic)))

	return mem.NewReader(pic[offset:end]), nil
}

func (repo *Repository) Scan(ctx *astral.Context, follow bool) (<-chan *astral.ObjectID, error) {
	ch := make(chan *astral.ObjectID)

	var subscribe <-chan *astral.ObjectID

	go func() {
		defer close(ch)

		if follow {
			subscribe = repo.scanQueue.Subscribe(ctx)
		}

		ids, err := repo.mod.db.UniquePictureIDs()
		if err != nil {
			repo.mod.log.Error("db error: %v", err)
			return
		}

		for _, id := range ids {
			select {
			case ch <- id:
			case <-ctx.Done():
				return
			}
		}

		// handle subscription
		if subscribe != nil {
			for id := range subscribe {
				select {
				case ch <- id:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return ch, nil
}

func (repo *Repository) String() string {
	return repo.Label()
}

func (repo *Repository) push(id *astral.ObjectID) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	repo.scanQueue = repo.scanQueue.Push(id)
}
