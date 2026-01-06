package objects

import (
	"errors"
	"io"
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/sig"
)

type RepoGroup struct {
	Concurrent bool
	mod        *Module
	label      string
	repos      sig.Set[string]
}

func NewRepoGroup(mod *Module, label string, concurrent bool) *RepoGroup {
	return &RepoGroup{
		mod:        mod,
		label:      label,
		Concurrent: concurrent,
	}
}

func (group *RepoGroup) Label() string {
	return group.label
}

func (group *RepoGroup) Create(ctx *astral.Context, opts *objects.CreateOpts) (objects.Writer, error) {
	var errs []error

	for _, repoName := range group.repos.Clone() {
		repo := group.mod.GetRepository(repoName)
		if repo == nil {
			continue
		}
		w, err := repo.Create(ctx, opts)
		if err == nil {
			return w, nil
		}
		errs = append(errs, err)
	}

	if len(errs) == 0 {
		return nil, errors.New("repository group empty")
	}

	return nil, errors.Join(errs...)
}

func (group *RepoGroup) Contains(ctx *astral.Context, objectID *astral.ObjectID) (bool, error) {
	for _, repoName := range group.repos.Clone() {
		repo := group.mod.GetRepository(repoName)
		if repo == nil {
			continue
		}
		if has, err := repo.Contains(ctx, objectID); err == nil && has {
			return true, nil
		}
	}
	return false, nil
}

func (group *RepoGroup) Scan(ctx *astral.Context, follow bool) (<-chan *astral.ObjectID, error) {
	ch := make(chan *astral.ObjectID)

	go func() {
		var wg sync.WaitGroup

		for _, repoName := range group.repos.Clone() {
			repo := group.mod.GetRepository(repoName)
			if repo == nil {
				continue
			}

			wg.Add(1)
			go func() {
				defer wg.Done()

				sub, err := repo.Scan(ctx, follow)
				if err != nil {
					return
				}

				var id *astral.ObjectID
				var ok bool

				// copy all scanned ids
				for {
					// read
					select {
					case <-ctx.Done():
						return
					case id, ok = <-sub:
						if !ok {
							return
						}
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

func (group *RepoGroup) Delete(ctx *astral.Context, objectID *astral.ObjectID) error {
	var count int
	for _, repoName := range group.repos.Clone() {
		repo := group.mod.GetRepository(repoName)
		if repo == nil {
			continue
		}
		if err := repo.Delete(ctx, objectID); err == nil {
			count++
		}
	}
	if count == 0 {
		return objects.ErrNotFound
	}
	return nil
}

func (group *RepoGroup) Read(ctx *astral.Context, objectID *astral.ObjectID, offset int64, limit int64) (io.ReadCloser, error) {
	if group.Concurrent {
		return group.readConcurrent(ctx, objectID, offset, limit)
	}

	return group.readSeq(ctx, objectID, offset, limit)
}

func (group *RepoGroup) readSeq(ctx *astral.Context, objectID *astral.ObjectID, offset int64, limit int64) (io.ReadCloser, error) {
	for _, repoName := range group.repos.Clone() {
		repo := group.mod.GetRepository(repoName)
		if repo == nil {
			continue
		}
		r, err := repo.Read(ctx, objectID, offset, limit)
		if err == nil {
			return r, nil
		}
	}
	return nil, objects.ErrNotFound
}

func (group *RepoGroup) readConcurrent(ctx *astral.Context, objectID *astral.ObjectID, offset int64, limit int64) (io.ReadCloser, error) {
	ctx, cancel := ctx.WithCancel()
	defer cancel()

	var res = make(chan io.ReadCloser)
	var wg sync.WaitGroup

	for _, repoName := range group.repos.Clone() {
		repo := group.mod.GetRepository(repoName)
		if repo == nil {
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			conn, err := repo.Read(ctx, objectID, offset, limit)
			if err == nil {
				select {
				case res <- conn:
				case <-ctx.Done():
				}
				return
			}
		}()
	}

	go func() {
		wg.Wait()
		close(res)
	}()

	select {
	case r, ok := <-res:
		if ok {
			return r, nil
		}

		return nil, objects.ErrNotFound

	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (group *RepoGroup) Free(ctx *astral.Context) (int64, error) {
	var total int64

	for _, repoName := range group.repos.Clone() {
		repo := group.mod.GetRepository(repoName)
		if repo == nil {
			continue
		}
		size, err := repo.Free(ctx)
		if err != nil {
			return 0, err
		}
		if size > 0 { // size might be -1 for unknown
			total += size
		}
	}

	return total, nil
}

func (group *RepoGroup) Add(repo string) error {
	if group.mod.GetRepository(repo) == nil {
		return errors.New("repository " + repo + " not found")
	}

	return group.repos.Add(repo)
}

func (group *RepoGroup) Remove(repo string) error {
	return group.repos.Remove(repo)
}

func (group *RepoGroup) List() []string {
	return group.repos.Clone()
}

func (group *RepoGroup) String() string {
	return group.label
}
