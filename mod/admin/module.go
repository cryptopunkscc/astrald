package admin

import (
	"bitbucket.org/creachadair/shell"
	"context"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/modules"
	"io"
)

var _ modules.Module = &Module{}

type Module struct {
	config   Config
	node     node.Node
	commands map[string]Command
	log      *log.Logger
}

func (mod *Module) Run(ctx context.Context) error {
	port, err := mod.node.Services().RegisterContext(ctx, "admin")
	if err != nil {
		return err
	}

	for query := range port.Queries() {
		// Only accept local queries
		if !query.IsLocal() {
			query.Reject()
			continue
		}

		conn, err := query.Accept()
		if err != nil {
			mod.log.Errorv(2, "accept: %s", err)
			continue
		}

		go func() {
			if err := mod.serve(conn, mod.node); err != nil {
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

		args, valid := shell.Split(line)
		if len(args) == 0 {
			continue
		}
		if !valid {
			term.Printf("error: unclosed quotes\n")
			continue
		}

		if cmd, found := mod.commands[args[0]]; found {
			err := cmd.Exec(term, args)
			if err != nil {
				term.Printf("error: %v\n", err)
			} else {
				term.Printf("ok\n")
			}
		} else {
			term.Printf("command not found\n")
		}
	}
}
