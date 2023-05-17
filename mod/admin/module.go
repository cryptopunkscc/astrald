package admin

import (
	"bitbucket.org/creachadair/shell"
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/modules"
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
