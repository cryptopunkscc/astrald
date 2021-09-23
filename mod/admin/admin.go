package admin

import (
	"bufio"
	"context"
	"fmt"
	_node "github.com/cryptopunkscc/astrald/node"
	"io"
	"strings"
)

type cmdFunc func(io.ReadWriter, *_node.Node, []string) error
type cmdMap map[string]cmdFunc

var commands cmdMap

func help(stream io.ReadWriter, _ *_node.Node, _ []string) error {
	fmt.Fprintf(stream, "commands:")
	for k := range commands {
		fmt.Fprintf(stream, " %s", k)
	}
	fmt.Fprintf(stream, "\n")

	return nil
}

func serve(stream io.ReadWriteCloser, node *_node.Node) {
	scanner := bufio.NewScanner(stream)

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
			fmt.Fprintf(stream, "no such command\n")
		}

	}
	stream.Close()
}

func listen(ctx context.Context, node *_node.Node) error {
	port, err := node.Hub.Register("admin")
	if err != nil {
		return err
	}

	for req := range port.Requests() {
		// Only accept local requests
		if req.Caller() != node.Network.Identity() {
			req.Reject()
			continue
		}
		conn := req.Accept()

		go serve(conn, node)
	}

	return nil
}

func init() {
	commands = cmdMap{
		"help":  help,
		"links": links,
	}
	_ = _node.RegisterService("admin", listen)
}
