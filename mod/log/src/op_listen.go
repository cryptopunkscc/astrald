package log

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opListenArgs struct {
	In  string `query:"optional"`
	Out string `query:"optional"`
}

func (mod *Module) OpListen(ctx *astral.Context, q shell.Query, args opListenArgs) (err error) {
	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	fwd := &logForwarder{ch}

	mod.outputs.Add(fwd)
	defer mod.outputs.Remove(fwd)

	for {
		_, err := ch.Receive()
		if err != nil {
			return nil
		}
	}
}

type logForwarder struct {
	ch channel.Sender
}

func (l logForwarder) LogEntry(entry *log.Entry) {
	l.ch.Send(entry)
}
