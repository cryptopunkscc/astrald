package core

import (
	"github.com/cryptopunkscc/astrald/lib/arl"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/object"
	"io"
	_log "log"
	"os"
	"strings"
	"time"
)

type logFields struct {
	screenOutput *log.LinePrinter
	screenFilter *log.PrinterFilter
	logOutput    *log.PrinterSplitter
	log          *log.Logger
}

func (node *Node) PushFormatFunc(fn log.FormatFunc) {
	node.log.PushFormatFunc(fn)
}

func (node *Node) setupLogs() {
	//TODO: output to our native log
	_log.SetOutput(io.Discard)

	node.screenOutput = log.NewLinePrinter(log.NewColorOutput(os.Stdout))
	node.screenFilter = log.NewPrinterFilter(node.screenOutput)
	node.logOutput = log.NewPrinterSplitter(node.screenFilter)
	node.log = log.NewLogger(node.logOutput).Tag(logTag)

	// string format
	node.log.PushFormatFunc(func(v any) ([]log.Op, bool) {
		s, ok := v.(string)
		if !ok {
			return nil, false
		}
		return []log.Op{
			log.OpColor{Color: log.Yellow},
			log.OpText{Text: s},
			log.OpReset{},
		}, true
	})

	// error format
	node.log.PushFormatFunc(func(v any) ([]log.Op, bool) {
		s, ok := v.(error)
		if !ok {
			return nil, false
		}
		return []log.Op{
			log.OpColor{Color: log.Red},
			log.OpText{Text: s.Error()},
			log.OpReset{},
		}, true
	})

	// time.Duration format
	node.log.PushFormatFunc(func(v any) ([]log.Op, bool) {
		s, ok := v.(time.Duration)
		if !ok {
			return nil, false
		}
		return []log.Op{
			log.OpColor{Color: log.Magenta},
			log.OpText{Text: s.String()},
			log.OpReset{},
		}, true
	})

	// ARL format
	node.log.PushFormatFunc(func(v any) ([]log.Op, bool) {
		var ops []log.Op

		a, ok := v.(*arl.ARL)
		if !ok {
			return nil, false
		}

		if !a.Caller.IsZero() {
			if cops, b := node.log.Render(a.Caller); b {
				ops = append(ops, cops...)
				ops = append(ops, log.OpText{Text: "@"})
			} else {
				return nil, false
			}
		}

		if tops, b := node.log.Render(a.Target); b {
			ops = append(ops, tops...)
		} else {
			return nil, false
		}

		if len(a.Query) > 0 {
			if qops, b := node.log.Render(a.Query); b {
				ops = append(ops, log.OpText{Text: ":"})
				ops = append(ops, qops...)
			} else {
				return nil, false
			}
		}

		return ops, true
	})

	node.log.PushFormatFunc(func(v any) ([]log.Op, bool) {
		s, ok := v.(astral.Nonce)
		if !ok {
			return nil, false
		}
		return []log.Op{
			log.OpColor{Color: log.Yellow},
			log.OpText{Text: s.String()},
			log.OpReset{},
		}, true
	})

	node.log.PushFormatFunc(func(v any) ([]log.Op, bool) {
		dataID, ok := v.(object.ID)
		if !ok {
			return nil, false
		}

		var ops []log.Op
		var s = dataID.String()

		if strings.HasPrefix(s, "id1") {
			s = s[3:]
			ops = append(ops,
				log.OpColor{Color: log.Blue},
				log.OpText{Text: "id1"},
				log.OpReset{},
			)
		}

		ops = append(ops,
			log.OpColor{Color: log.BrightBlue},
			log.OpText{Text: s},
			log.OpReset{},
		)

		return ops, true
	})

	// nil format
	node.log.PushFormatFunc(func(v any) ([]log.Op, bool) {
		if v != nil {
			return nil, false
		}
		return []log.Op{
			log.OpColor{Color: log.Red},
			log.OpText{Text: "nil"},
			log.OpReset{},
		}, true
	})
}

func (node *Node) loadLogConfig() error {
	node.logConfig = defaultLogConfig

	if err := node.assets.LoadYAML("log", &node.logConfig); err != nil {
		return nil
	}

	for tag, level := range node.logConfig.TagLevels {
		node.screenFilter.TagLevels[tag] = level
	}
	for tag, color := range node.logConfig.TagColors {
		node.screenOutput.TagColors[tag] = log.ParseColor(color)
	}
	node.screenOutput.SetHideDate(node.logConfig.HideDate)
	node.screenFilter.Level = node.logConfig.Level

	var logFile = node.logConfig.LogFile

	if logFile != "" {
		fileOutput, err := log.NewFileOutput(logFile)
		if err != nil {
			node.log.Error("error opening log file %v: %v", logFile, err)
			return nil
		}

		filePrinter := log.NewLinePrinter(fileOutput)
		node.logOutput.Add(filePrinter)
	}

	return nil
}
