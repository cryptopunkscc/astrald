package log

import (
	"time"
)

type PrinterSplitter struct {
	Printers []Printer
}

func NewPrinterSplitter(printers ...Printer) *PrinterSplitter {
	return &PrinterSplitter{Printers: printers}
}

func (p *PrinterSplitter) Log(t Type, level int, ts time.Time, tag string, ops ...Op) {
	for _, printer := range p.Printers {
		printer.Log(t, level, ts, tag, ops...)
	}
}

func (p *PrinterSplitter) Add(printer Printer) {
	p.Printers = append(p.Printers, printer)
}
