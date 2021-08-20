package test

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/node"
)

const port = "lore-test"
const testStoryType = "test_type"
const testStoryAuthor = "test_author"

func init() {
	_ = node.RegisterService(port, run)
}

func run(ctx context.Context, core api.Core) (err error) {
	go func() { observeLore(ctx, core) }()
	//go func() { mapFile(ctx, core) }()
	//go func() { spamRepo(ctx, core) }()
	go func() { scanDir(ctx, core) }()

	<-ctx.Done()
	return nil
}
