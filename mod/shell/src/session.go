package shell

import (
	shell2 "bitbucket.org/creachadair/shell"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/term"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"io"
)

type Session struct {
	mod  *Module
	conn io.ReadWriter
	env  *shell.Env
}

func NewSession(mod *Module, conn io.ReadWriter) *Session {
	return &Session{
		mod:  mod,
		conn: conn,
		env:  shell.NewTextEnv(conn, conn),
	}
}

func (s *Session) Run(ctx astral.Context) error {
	var p = term.NewBasicPrinter(s.conn, &term.DefaultTypeMap)

	for {
		p.Print(&Prompt{
			guestID: ctx.Identitiy(),
			hostID:  s.mod.node.Identity(),
		})

		obj, _, err := s.env.ReadObject()
		if err != nil {
			return err
		}
		if obj == nil {
			return nil
		}

		var line string

		if s, ok := obj.(fmt.Stringer); ok {
			line = s.String()
		} else {
			return errors.New("unsupported object on input")
		}

		args, valid := shell2.Split(line)
		switch {
		case !valid:
			term.Printf(p, "quote mismatch\n")
			continue

		case len(args) == 0:
			continue
		}

		if args[0] == "exit" {
			return nil
		}

		err = s.mod.root.CallArgs(ctx, s.env, args[0], args[1:])
		if err != nil {
			term.Printf(p, "error: %v\n", err)
		} else {
			term.Printf(p, "ok\n")
		}
	}
}
