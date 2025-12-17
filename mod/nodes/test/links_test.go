package nodes

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/astral/term"
	"github.com/cryptopunkscc/astrald/core/assets"
	nodes "github.com/cryptopunkscc/astrald/mod/nodes/src"
	"github.com/cryptopunkscc/astrald/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ConnectAccept(t *testing.T) {
	n1, m1 := newTestModule(t)
	n2, m2 := newTestModule(t)
	inConn, outConn := PipeConn(
		&Endpoint{address: "1"},
		&Endpoint{address: "2"},
	)

	ctx := context.Background()
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		t.Log("verify connect")
		assert.NoError(t, m1.Connect(ctx, n2.Identity(), outConn))
		assert.True(t, m1.IsLinked(n2.Identity()))

		t.Log("done 1")
		wg.Done()
	}()
	go func() {
		t.Log("verify accept")
		assert.NoError(t, m2.Accept(ctx, inConn))
		assert.True(t, m2.IsLinked(n1.Identity()))

		t.Log("done 2")
		wg.Done()
	}()
	wg.Wait()
}

func newTestModule(t *testing.T) (astral.Node, *nodes.Module) {
	node, err := NewNode()
	require.NoError(t, err)

	printer := term.NewBasicPrinter(os.Stdout, &term.DefaultTypeMap)
	logger := log.NewLogger(printer, node.Identity(), "test")

	memResources := resources.NewMemResources()
	coreAssets, err := assets.NewCoreAssets(memResources, logger)
	require.NoError(t, err)

	loaded, err := nodes.Loader{}.Load(node, coreAssets, logger)
	require.NoError(t, err)

	module := loaded.(*nodes.Module)
	require.NotNil(t, module)

	module.Events = &EventsModule{t}
	module.Objects = &ObjectsModule{t}
	return node, module
}
