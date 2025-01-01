package shell

import (
	shell2 "bitbucket.org/creachadair/shell"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"io"
)

type Session struct {
	mod  *Module
	conn io.ReadWriteCloser
}

func NewSession(mod *Module, conn io.ReadWriteCloser) *Session {
	return &Session{
		mod:  mod,
		conn: conn,
	}
}

func (s *Session) Run(ctx astral.Context) (err error) {
	var t = shell.NewTerminal(s.conn)
	
	for {
		// print the prompt
		t.Print(&Prompt{
			guestID: ctx.Identity(),
			hostID:  s.mod.node.Identity(),
		})

		// read the command
		var line string
		line, err = t.ReadLine()
		if err != nil {
			return err
		}

		args, valid := shell2.Split(line)
		switch {
		case !valid:
			t.Printf("quote mismatch\n")
			continue

		case len(args) == 0:
			continue
		}

		op := args[0]
		if op == "exit" {
			return nil
		}

		params := shell.ParseArgs(args[1:])

		var q = query.New(ctx.Identity(), s.mod.node.Identity(), op, params)

		conn, err := query.Route(ctx, s.mod.node, q)
		if err != nil {
			t.Printf("error: %v\n", err)
			continue
		}

		//TODO: forward input from the user to conn

		io.Copy(s.conn, conn)
		conn.Close()

		t.Printf("ok\n")
	}
}
