package indexing

import (
	"context"
	"errors"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/indexing"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/mod/tree"
	"github.com/cryptopunkscc/astrald/resources"
	"github.com/cryptopunkscc/astrald/sig"
)

type Deps struct {
	Objects objects.Module
	Tree    tree.Module
}

type Module struct {
	Deps
	config  Config
	node    astral.Node
	log     *log.Logger
	assets  resources.Resources
	ops     shell.Scope
	db      *DB
	ctx     *astral.Context
	repos   tree.Node
	indexes tree.Node

	mod sig.Map[string, context.CancelFunc]
}

func (mod *Module) Run(ctx *astral.Context) error {
	mod.ctx = ctx

	sub, err := mod.repos.Sub(ctx)
	if err != nil {
		return err
	}

	for repoName := range sub {
		err = mod.startRepoSync(repoName)
		if err != nil {
			mod.log.Logv(1, "error starting repo sync: %v", err)
		}
	}

	<-ctx.Done()
	return nil
}

func (mod *Module) Scope() *shell.Scope {
	return &mod.ops
}

func (mod *Module) String() string {
	return indexing.ModuleName
}

func (mod *Module) EnableRepo(ctx *astral.Context, repoName string) error {
	_, err := mod.repos.Create(ctx, repoName)
	if err != nil {
		return err
	}

	return mod.startRepoSync(repoName)
}

func (mod *Module) DisableRepo(ctx *astral.Context, repoName string) error {
	mod.stopRepoSync(repoName)

	sub, err := mod.repos.Sub(ctx)
	if err != nil {
		return err
	}

	del, ok := sub[repoName]
	if ok {
		del.Delete(ctx)
	}

	return nil
}

func (mod *Module) CreateIndex(ctx *astral.Context, indexName string) error {
	indexNode, err := mod.indexes.Create(ctx, indexName)
	if err != nil {
		return err
	}

	var height = astral.Uint64(0)

	indexNode.Set(ctx, &height)

	return nil
}

func (mod *Module) startRepoSync(repoName string) error {
	ctx, cancel := mod.ctx.WithCancel()

	_, ok := mod.mod.Set(repoName, cancel)
	if !ok {
		cancel()
		return errors.New("repo already syncing")
	}

	go func() {
		mod.log.Logv(1, "following repo %v", repoName)
		err := mod.syncRepo(ctx, repoName)
		if err != nil {
			mod.log.Logv(1, "error syncing repo %v: %v", repoName, err)
		} else {
			mod.log.Logv(1, "stopped following %v", repoName)
		}
	}()

	return nil
}

func (mod *Module) stopRepoSync(repoName string) error {
	cancel, ok := mod.mod.Delete(repoName)
	if !ok {
		return errors.New("repo not syncing")
	}
	cancel()
	return nil
}

func (mod *Module) syncRepo(ctx *astral.Context, repoName string) error {
	repo := mod.Objects.GetRepository(repoName)
	if repo == nil {
		return errors.New("repository not found: " + repoName)
	}

	scan, err := repo.Scan(ctx, true)
	if err != nil {
		return err
	}

	// take a snapshot
	var snapshot []*astral.ObjectID

	timeout := time.NewTimer(time.Second * 1)
	defer timeout.Stop()

	for {
		select {
		case objectJD := <-scan:
			if objectJD == nil {
				goto snapshot
			}
			snapshot = append(snapshot, objectJD)
			if !timeout.Stop() {
				<-timeout.C
			}
			timeout.Reset(time.Second * 1)
		case <-timeout.C:
			goto snapshot
		}
	}

snapshot:

	var removed, added int

	// remove objects from the index that are not in the snapshot
	excess, err := mod.db.findExcessObjectIDs(repoName, snapshot)
	if err != nil {
		return err
	}

	for _, objectID := range excess {
		removed++
		err = mod.db.removeFromRepo(repoName, objectID)
		if err != nil {
			mod.log.Logv(1, "db error removing from repo: %v", err)
		}

		// check context
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}

	// add objects from the snapshot that are not in the index
	missing, err := mod.db.findMissingObjectIDs(repoName, snapshot)
	for _, objectID := range missing {
		added++
		err = mod.db.addToRepo(repoName, objectID)
		if err != nil {
			mod.log.Logv(1, "error adding %v to repo %v: %v", objectID, repoName, err)
		}

		// check context
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}

	mod.log.Logv(1, "index synced with repo %v: %v removed, %v added.", repoName, removed, added)

	// follow updates from the repo until ctx is canceled
	for objectID := range scan {
		err = mod.db.addToRepo(repoName, objectID)
		if err != nil {
			mod.log.Logv(1, "db error adding to repo: %v", err)
		}
	}

	return nil
}
