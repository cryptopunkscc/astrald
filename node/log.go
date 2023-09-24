package node

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/assets"
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

func (node *CoreNode) setupLogs() {
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

	// id.Identity format
	node.log.PushFormatFunc(func(v any) ([]log.Op, bool) {
		identity, ok := v.(id.Identity)
		if !ok {
			return nil, false
		}

		var color = log.Cyan

		if ks, err := node.assets.KeyStore(); err == nil {
			if identity, err := ks.Find(identity); err == nil {
				if identity.PrivateKey() != nil {
					color = log.Green
				}
			}
		}

		if node.identity.IsEqual(identity) {
			color = log.BrightGreen
		}

		var name = node.Resolver().DisplayName(identity)

		return []log.Op{
			log.OpColor{Color: color},
			log.OpText{Text: name},
			log.OpReset{},
		}, true
	})

	node.log.PushFormatFunc(func(v any) ([]log.Op, bool) {
		s, ok := v.(net.Nonce)
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
		dataID, ok := v.(data.ID)
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

func (node *CoreNode) loadLogConfig(assets assets.Store) error {
	if err := assets.LoadYAML("log", &node.logConfig); err != nil {
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
