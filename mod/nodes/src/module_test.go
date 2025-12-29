package nodes

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/astral/term"
	"github.com/cryptopunkscc/astrald/core/assets"
	test "github.com/cryptopunkscc/astrald/mod/nodes/src/test"
	"github.com/cryptopunkscc/astrald/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModule_ConnectAccept(t *testing.T) {
	module1 := newTestModule(t)
	module2 := newTestModule(t)
	conn1, conn2 := test.PipeConn(&test.Endpoint{Addr: "1"}, &test.Endpoint{Addr: "2"})

	t.Run("connect", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		assert.NoError(t, module1.Connect(ctx, module2.node.Identity(), conn1))
		assert.True(t, module1.IsLinked(module2.node.Identity()))
	})

	t.Run("accept", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		assert.NoError(t, module2.Accept(ctx, conn2))
		assert.True(t, module2.IsLinked(module1.node.Identity()))
	})
}

func newTestModule(t *testing.T) *Module {
	node := test.NewNode()

	printer := term.NewBasicPrinter(os.Stdout, &term.DefaultTypeMap)
	logger := log.NewLogger(printer, node.Identity(), "test")

	memResources := resources.NewMemResources()
	coreAssets, err := assets.NewCoreAssets(memResources, logger)
	require.NoError(t, err)

	loaded, err := Loader{}.Load(node, coreAssets, logger)
	require.NoError(t, err)

	module := loaded.(*Module)
	require.NotNil(t, module)

	module.Events = &test.EventsModule{T: t}
	module.Objects = &test.ObjectsModule{T: t}
	return module
}
