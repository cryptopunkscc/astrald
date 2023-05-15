package admin

import (
	"bitbucket.org/creachadair/shell"
	"bufio"
	"context"
	"fmt"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/modules"
	"io"
)

var _ modules.Module = &Module{}

type Module struct {
	config   Config
	node     node.Node
	commands map[string]Command
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

		if err == nil {
			go mod.serve(conn, mod.node)
		}
	}

	return nil
}

func (mod *Module) serve(stream io.ReadWriteCloser, node node.Node) {
	defer stream.Close()

	prompt := node.Alias() + mod.config.Prompt

	scanner := bufio.NewScanner(stream)
	stream.Write([]byte(prompt))

	for scanner.Scan() {
		split, valid := shell.Split(scanner.Text())
		if len(split) == 0 {
			goto prompt
		}
		if !valid {
			fmt.Fprintf(stream, "error: unclosed quotes\n")
			goto prompt
		}

		if c, found := mod.commands[split[0]]; found {
			err := c.Exec(NewTerminal(stream), split)
			if err != nil {
				fmt.Fprintf(stream, "error: %v\n", err)
			} else {
				fmt.Fprintf(stream, "ok\n")
			}
		} else {
			fn, ok := commands[split[0]]
			if ok {
				err := fn(stream, node, split[1:])
				if err != nil {
					fmt.Fprintf(stream, "error: %v\n", err)
				} else {
					fmt.Fprintf(stream, "ok\n")
				}
			} else {
				if len(split[0]) > 0 {
					fmt.Fprintf(stream, "no such command\n")
				}
			}
		}

	prompt:
		prompt = node.Alias() + mod.config.Prompt
		stream.Write([]byte(prompt))
	}
}
