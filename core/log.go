package core

import (
	"bytes"
	"errors"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/astral/term"
	"io"
	_log "log"
	"os"
)

const maxAdapterBuffer = 1 << 16 // 64kb

type logFields struct {
	log     *log.Logger
	printer *term.BasicPrinter
}

type adapter struct {
	*log.Logger
	b []byte
}

var _ io.Writer = &adapter{}

func (node *Node) initLogger() {
	node.printer = term.NewBasicPrinter(os.Stdout, &term.DefaultTypeMap)
	node.printer.Mono = true

	node.log = log.NewLogger(node.printer, node.identity, logTag)
	node.log.Level = 2

	_log.SetOutput(adapter{Logger: node.log.Tag("log")})
	_log.SetFlags(0)
}

func (node *Node) configureLogger() {
	if !node.config.Log.DisableColors {
		node.printer.Mono = false
	}

	node.log.Level = node.config.Log.Level
}

func (a adapter) Write(p []byte) (n int, err error) {
	if len(a.b)+len(p) > maxAdapterBuffer {
		return 0, errors.New("log buffer overflow")
	}

	a.b = append(a.b, p...)
	n = len(p)

	idx := bytes.IndexByte(a.b, '\n')
	if idx == -1 {
		return
	}

	line := a.b[:idx]
	a.b = a.b[idx+1:]

	a.Logger.Log("%v", string(line))

	return
}
