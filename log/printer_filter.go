package log

import "time"

type PrinterFilter struct {
	Printer
	Level     int
	TagLevels map[string]int
}

func NewPrinterFilter(printer Printer) *PrinterFilter {
	return &PrinterFilter{
		Printer:   printer,
		TagLevels: map[string]int{},
	}
}

func (p *PrinterFilter) Log(t Type, level int, ts time.Time, tag string, ops ...Op) {
	if p.check(t, level, ts, tag, ops...) {
		p.Printer.Log(t, level, ts, tag, ops...)
	}
}

func (p *PrinterFilter) check(t Type, level int, ts time.Time, tag string, ops ...Op) bool {
	if l, ok := p.TagLevels[tag]; ok {
		return level <= l
	}

	return level <= p.Level
}
