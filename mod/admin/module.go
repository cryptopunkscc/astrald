package admin

import (
	"bitbucket.org/creachadair/shell"
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/modules"
	"github.com/cryptopunkscc/astrald/node/services"
	"io"
	"strings"
)

var _ modules.Module = &Module{}

type Module struct {
	config   Config
	node     node.Node
	commands map[string]Command
	log      *log.Logger
}

func (mod *Module) Run(ctx context.Context) error {
	var queries = services.NewQueryChan(4)
	service, err := mod.node.Services().Register(ctx, mod.node.Identity(), "admin", queries.Push)
	if err != nil {
		return err
	}

	go func() {
		<-service.Done()
		close(queries)
	}()

	for query := range queries {
		// Only accept local queries
		if query.Origin() != services.OriginLocal {
			query.Reject()
			continue
		}

		conn, err := query.Accept()
		if err != nil {
			mod.log.Errorv(2, "accept: %s", err)
			continue
		}

		go func() {
			err := mod.serve(conn, mod.node)
			switch {
			case err == nil:
			case strings.Contains(err.Error(), "on closed pipe"):
			default:
				mod.log.Errorv(2, "serve: %s", err)
			}
		}()
	}

	return nil
}

func (mod *Module) serve(stream io.ReadWriteCloser, node node.Node) error {
	defer stream.Close()

	var term = NewTerminal(stream)

	for {
		term.Printf("%s%s", node.Alias(), mod.config.Prompt)

		line, err := term.ScanLine()
		if err != nil {
			return err
		}

		if err := mod.exec(line, term); err != nil {
			term.Printf("error: %v\n", err)
		} else {
			term.Printf("ok\n")
		}
	}
}

func (mod *Module) exec(line string, term *Terminal) error {
	args, valid := shell.Split(line)
	if len(args) == 0 {
		return nil
	}
	if !valid {
		return errors.New("unclosed quotes")
	}

	if cmd, found := mod.commands[args[0]]; found {
		return cmd.Exec(term, args)
	} else {
		return errors.New("command not found")
	}
}
