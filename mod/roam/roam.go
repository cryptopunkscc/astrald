package roam

import (
	"context"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/tasks"
)

const (
	pickServiceName = "roam.pick"
	dropServiceName = "roam.drop"
)

type Module struct {
	node  node.Node
	moves map[int]*node.Conn
	log   *log.Logger
	ctx   context.Context
}

func (mod *Module) Run(ctx context.Context) error {
	mod.ctx = ctx
	mod.moves = make(map[int]*node.Conn)

	return tasks.Group(
		&PickService{Module: mod},
		&DropService{Module: mod},
		&OptimizerService{Module: mod},
	).Run(ctx)
}

func (mod *Module) allocMove(conn *node.Conn) int {
	id := mod.unusedMoveID()
	if id != -1 {
		mod.moves[id] = conn
	}
	return id
}

func (mod *Module) unusedMoveID() int {
	for i := 0; i < 255; i++ {
		if _, ok := mod.moves[i]; !ok {
			return i
		}
	}
	return -1
}
