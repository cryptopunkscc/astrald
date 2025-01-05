package core

import (
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/astral/term"
	"io"
	_log "log"
	"os"
)

type logFields struct {
	log     *log.Logger
	printer term.Printer
}

func (node *Node) setupLogs() {
	node.printer = term.NewBasicPrinter(os.Stdout, &term.DefaultTypeMap)

	//TODO: output to our native log
	_log.SetOutput(io.Discard)

	node.log = log.NewLogger(node.printer, node.identity, logTag)
	node.log.Level = 2
}

func (node *Node) loadLogConfig() error {
	node.logConfig = defaultLogConfig

	if err := node.assets.LoadYAML("log", &node.logConfig); err != nil {
		return nil
	}

	return nil
}
