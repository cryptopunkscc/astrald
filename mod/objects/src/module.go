package objects

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/sig"
)

var _ objects.Module = &Module{}

type Deps struct {
	Auth  auth.Module
	Dir   dir.Module
	Nodes nodes.Module
}

type Module struct {
	Deps
	node   astral.Node
	config Config
	db     *DB
	log    *log.Logger
	ops    shell.Scope

	ctx        *astral.Context
	describers sig.Set[objects.Describer]
	searchers  sig.Set[objects.Searcher]
	searchPre  sig.Set[objects.SearchPreprocessor]
	finders    sig.Set[objects.Finder]
	receivers  sig.Set[objects.Receiver]
	holders    sig.Set[objects.Holder]
	repos      sig.Map[string, objects.Repository]

	groups sig.Map[string, *RepoGroup]
}

func (mod *Module) Probe(ctx *astral.Context, repo objects.Repository, objectID *astral.ObjectID) (probe *objects.Probe, err error) {
	probe = &objects.Probe{}

	startAt := time.Now()

	// read the object data
	r, err := repo.Read(ctx, objectID, 0, 512)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	// store the response time
	probe.Time = astral.Duration(time.Since(startAt))

	// store the actual repo name
	probe.Repo = astral.String8(mod.getRepoName(r.Repo()))

	// check if it's an astral object
	q := bytes.NewReader(data)
	if _, err := (&astral.Stamp{}).ReadFrom(q); err == nil {
		var t astral.ObjectType
		_, err = t.ReadFrom(q)
		if err == nil {
			probe.Type = astral.String8(t.String())
		}
	}

	// check the mimeType
	probe.Mime = astral.String8(http.DetectContentType(data))

	return
}

func (mod *Module) Run(ctx *astral.Context) error {
	mod.ctx = ctx

	<-ctx.Done()

	return nil
}

func (mod *Module) Load(ctx *astral.Context, repo objects.Repository, objectID *astral.ObjectID) (astral.Object, error) {
	// read the object data
	r, err := repo.Read(ctx, objectID, 0, 0)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	// parse the object
	o, _, err := astral.Decode(bytes.NewReader(data), astral.Canonical())
	switch {
	case err == nil:
		return o, nil // object successfully loaded

	case strings.Contains(err.Error(), "invalid magic bytes"): // the object is a blob
		return (*astral.Blob)(&data), nil

	default: // other error
		return nil, err
	}
}

func (mod *Module) Store(ctx *astral.Context, repo objects.Repository, object astral.Object) (*astral.ObjectID, error) {
	w, err := repo.Create(ctx, nil)
	if err != nil {
		return nil, err
	}

	_, err = astral.Encode(w, object, astral.WithEncoder(astral.CanonicalTypeEncoder))
	if err != nil {
		return nil, err
	}

	return w.Commit()
}

// Deprecated: Use Probe instead.
func (mod *Module) GetType(ctx *astral.Context, objectID *astral.ObjectID) (objectType string, err error) {
	// check the cache
	row, err := mod.db.Find(objectID)
	if err == nil {
		return row.Type, nil
	}

	// read first bytes of the object
	r, err := mod.ReadDefault().Read(ctx, objectID, 0, 260) // max header size: 4 magic bytes + 1 len + 255 type
	if err != nil {
		return "", objects.ErrNotFound
	}
	defer r.Close()

	// read the stamp
	_, err = (&astral.Stamp{}).ReadFrom(r)
	if err != nil {
		return "", errors.New("missing astral stamp")
	}

	// read the type
	var t astral.ObjectType
	_, err = t.ReadFrom(r)
	if err != nil {
		return "", err
	}

	// write to cache
	err = mod.db.Create(objectID, t.String())
	switch {
	case err == nil:
	case strings.Contains(err.Error(), "UNIQUE constraint failed"):
	default:
		mod.log.Error("onSave: db error: %v", err)
	}

	return t.String(), nil
}

func (mod *Module) AddRepository(name string, repo objects.Repository) error {
	_, ok := mod.repos.Set(name, repo)
	if !ok {
		return fmt.Errorf("repo %s already added", repo.Label())
	}
	return nil
}

func (mod *Module) GetRepository(name string) (repo objects.Repository) {
	repo, _ = mod.repos.Get(name)
	return
}

func (mod *Module) RemoveRepository(name string) error {
	if len(name) == 0 {
		return errors.New("name is empty")
	}

	removed, ok := mod.repos.Delete(name)
	if !ok {
		return fmt.Errorf("repository %s not found", name)
	}

	if c, ok := removed.(objects.AfterRemovedCallback); ok {
		c.AfterRemoved(name)
	}

	return nil
}

// ReadDefault returns the default repository for reading objects
func (mod *Module) ReadDefault() (repo objects.Repository) {
	return mod.GetRepository("main")
}

// WriteDefault returns the default repository for writing objects
func (mod *Module) WriteDefault() (repo objects.Repository) {
	return mod.GetRepository("local")
}

// AddGroup adds a repository to a group
func (mod *Module) AddGroup(groupName string, repoName string) error {
	maybeGroup := mod.GetRepository(groupName)
	if maybeGroup == nil {
		return fmt.Errorf("repo %s not found", groupName)
	}

	group, ok := maybeGroup.(*RepoGroup)
	if !ok {
		return fmt.Errorf("repo %s is not a group", groupName)
	}

	return group.Add(repoName)
}

// RemoveGroup removes a repository from a group
func (mod *Module) RemoveGroup(groupName string, repoName string) error {
	maybeGroup := mod.GetRepository(groupName)
	if maybeGroup == nil {
		return fmt.Errorf("repo %s not found", groupName)
	}

	group, ok := maybeGroup.(*RepoGroup)
	if !ok {
		return fmt.Errorf("repo %s is not a group", groupName)
	}

	return group.Remove(repoName)
}

func (mod *Module) AddSearchPreprocessor(pre objects.SearchPreprocessor) error {
	return mod.searchPre.Add(pre)
}

// getRepoName returns the name of a repository
func (mod *Module) getRepoName(repo objects.Repository) string {
	for name, r := range mod.repos.Clone() {
		if r == repo {
			return name
		}
	}
	return ""
}

func (mod *Module) Scope() *shell.Scope {
	return &mod.ops
}

func (mod *Module) String() string {
	return objects.ModuleName
}
