package admin

import (
	"bufio"
	"context"
	"fmt"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/modules"
	"io"
	"strings"
)

const promptString = "> "

var _ modules.Module = &Admin{}

type Admin struct {
	node node.Node
}

type cmdFunc func(io.ReadWriter, node.Node, []string) error
type cmdMap map[string]cmdFunc

var commands cmdMap

func (mod *Admin) Run(ctx context.Context) error {
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
			go serve(conn, mod.node)
		}
	}

	return nil
}

func help(stream io.ReadWriter, _ node.Node, _ []string) error {
	fmt.Fprintf(stream, "commands:")
	for k := range commands {
		fmt.Fprintf(stream, " %s", k)
	}
	fmt.Fprintf(stream, "\n")

	return nil
}

func serve(stream io.ReadWriteCloser, node node.Node) {
	defer stream.Close()

	prompt := node.Alias() + promptString

	scanner := bufio.NewScanner(stream)
	stream.Write([]byte(prompt))

	for scanner.Scan() {
		words := strings.Split(scanner.Text(), " ")
		if len(words) == 0 {
			continue
		}

		cmd, args := words[0], words[1:]

		fn, ok := commands[cmd]
		if ok {
			err := fn(stream, node, args)
			if err != nil {
				fmt.Fprintf(stream, "error: %v\n", err)
			} else {
				fmt.Fprintf(stream, "ok\n")
			}
		} else {
			if len(cmd) > 0 {
				fmt.Fprintf(stream, "no such command\n")
			}
		}

		prompt = node.Alias() + promptString
		stream.Write([]byte(prompt))
	}
}

func init() {
	commands = cmdMap{
		"help":     help,
		"peers":    peers,
		"contacts": cmdContacts,
		"info":     info,
		"parse":    parse,
		"add":      add,
		"forget":   forget,
		"link":     link,
		"unlink":   unlink,
		"tracker":  cmdTracker,
		"check":    check,
		"launch":   launch,
	}
}
