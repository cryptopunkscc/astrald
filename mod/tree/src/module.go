package tree

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/lib/paths"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/mod/tree"
	"github.com/cryptopunkscc/astrald/resources"
	"github.com/cryptopunkscc/astrald/sig"
)

type Deps struct {
	Dir dir.Module
}

type Module struct {
	Deps
	config Config
	node   astral.Node
	log    *log.Logger
	assets resources.Resources
	ops    shell.Scope
	db     *DB

	// mounted nodes
	mounts sig.Map[string, tree.Node]

	// node value cache
	nodeValue   map[int]*sig.Queue[astral.Object]
	nodeValueMu sync.Mutex

	// active bindings by path
	bindings sig.Map[string, *sig.Set[*binding]]
}

var _ tree.Module = &Module{}

func (mod *Module) Run(ctx *astral.Context) error {
	<-ctx.Done()
	return nil
}

func (mod *Module) Get(ctx *astral.Context, path string) (astral.Object, error) {
	node, err := tree.Query(ctx, mod.Root(), path, false)
	if err != nil {
		return nil, err
	}

	get, err := node.Get(ctx, false)
	if err != nil {
		return nil, err
	}

	return <-get, nil
}

func (mod *Module) Set(ctx *astral.Context, path string, object astral.Object) error {
	node, err := tree.Query(ctx, mod.Root(), path, true)
	if err != nil {
		return err
	}

	return node.Set(ctx, object)
}

func (mod *Module) Delete(ctx *astral.Context, path string) error {
	node, err := tree.Query(ctx, mod.Root(), path, false)
	if err != nil {
		return err
	}

	if err := node.Delete(ctx); err != nil {
		return err
	}

	mod.invalidateBindings(path)
	return nil
}

func (mod *Module) Mount(path string, node tree.Node) error {
	// check and normalize the path
	if !strings.HasPrefix(path, "/") {
		return errors.New("path must be absolute")
	}
	path = strings.TrimSuffix(path, "/")

	// set the mount point
	_, ok := mod.mounts.Set(path, node)
	if !ok {
		return errors.New("mount point already exists")
	}

	return nil
}

func (mod *Module) Unmount(path string) error {
	// check and normalize the path
	if !strings.HasPrefix(path, "/") {
		return errors.New("path must be absolute")
	}
	path = strings.TrimSuffix(path, "/")

	// delete the mount point
	_, ok := mod.mounts.Delete(path)
	if !ok {
		return errors.New("mount point does not exist")
	}

	mod.invalidateBindings(path)
	return nil
}

func (mod *Module) Root() tree.Node {
	root, _ := mod.mounts.Get("/")
	if root == nil {
		return nil
	}

	return &NodeWrapper{
		Node: root,
		mod:  mod,
	}
}

func (mod *Module) Scope() *shell.Scope {
	return &mod.ops
}

func (mod *Module) String() string {
	return tree.ModuleName
}

// getMount returns the mounted node for the given path or nil if nothing is mounted.
func (mod *Module) getMount(path string) tree.Node {
	// check and normalize the path
	if !strings.HasPrefix(path, "/") {
		return nil
	}
	path = strings.TrimSuffix(path, "/")

	node, _ := mod.mounts.Get(path)

	return node
}

// pushNodeValue sets node's value cache and notifies all subscribers.
func (mod *Module) pushNodeValue(nodeID int, value astral.Object) {
	mod.nodeValueMu.Lock()
	defer mod.nodeValueMu.Unlock()

	queue, found := mod.nodeValue[nodeID]
	if !found {
		return // do nothing if there are no observers for this node
	}

	mod.nodeValue[nodeID] = queue.Push(value)
}

// subscribeNodeValue returns a channel that receives node's value changes until the context is canceled.
func (mod *Module) subscribeNodeValue(ctx context.Context, nodeID int) <-chan astral.Object {
	mod.nodeValueMu.Lock()
	queue, found := mod.nodeValue[nodeID]
	if !found {
		queue = &sig.Queue[astral.Object]{}
		mod.nodeValue[nodeID] = queue
	}
	mod.nodeValueMu.Unlock()

	var out = make(chan astral.Object)

	go func() {
		for v := range sig.Subscribe(ctx, queue) {
			select {
			case <-ctx.Done():
				return
			case out <- v:
			}
		}
	}()

	return out
}

// registerBinding adds a binding to the tracking map.
func (mod *Module) registerBinding(path string, b *binding) {
	set, ok := mod.bindings.Get(path)
	if !ok {
		set = &sig.Set[*binding]{}
		existing, added := mod.bindings.Set(path, set)
		if !added {
			set = existing
		}
	}
	set.Add(b)
}

// unregisterBinding removes a binding from the tracking map.
func (mod *Module) unregisterBinding(path string, b *binding) {
	set, ok := mod.bindings.Get(path)
	if !ok {
		return
	}
	set.Remove(b)
	if set.Count() == 0 {
		mod.bindings.Delete(path)
	}
}

// invalidateBindings closes all bindings at or under the given path.
func (mod *Module) invalidateBindings(path string) {
	var toClose []*binding

	// Close bindings at exact path
	set, ok := mod.bindings.Delete(path)
	if ok {
		for _, b := range set.Clone() {
			toClose = append(toClose, b)
		}
	}

	// Close bindings under the path
	mod.bindings.Each(func(p string, bindings *sig.Set[*binding]) error {
		if paths.PathUnder(p, path, '/') {
			for _, b := range bindings.Clone() {
				toClose = append(toClose, b)
			}
		}
		return nil
	})

	for _, b := range toClose {
		b.Close()
	}
}
