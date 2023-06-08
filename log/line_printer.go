package log

import (
	"fmt"
	"sync"
	"time"
)

var _ Printer = &LinePrinter{}

const defaultTagColor = White
const defaultTimeColor = White
const defaultLevelColor = White

type LinePrinter struct {
	TagColors  map[string]Color
	mu         sync.Mutex
	out        Output
	timeFormat string
	tagColor   Color
	timeColor  Color
	levelColor Color
}

func NewLinePrinter(output Output) *LinePrinter {
	return &LinePrinter{
		out:        output,
		timeFormat: shortTimestamp,
		tagColor:   defaultTagColor,
		timeColor:  defaultTimeColor,
		levelColor: defaultLevelColor,
		TagColors:  map[string]Color{},
	}
}

func (p *LinePrinter) SetTagColor(tagColor Color) {
	p.tagColor = tagColor
}

func (p *LinePrinter) SetTimeColor(timeColor Color) {
	p.timeColor = timeColor
}

func (p *LinePrinter) SetHideDate(s bool) {
	if s {
		p.timeFormat = shortTimestamp
	} else {
		p.timeFormat = fullTimestamp
	}
}

func (p *LinePrinter) Log(t Type, level int, ts time.Time, tag string, ops ...Op) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// level
	p.out.Do(
		OpColor{Color: p.getLevelColor(level)},
		OpText{Text: fmt.Sprintf("(%d)", level)},
		OpReset{},
		OpText{Text: " "},
	)

	// timestamp
	var timestamp = ts.Format(p.timeFormat)
	p.out.Do(
		OpColor{Color: p.timeColor},
		OpText{Text: timestamp},
		OpReset{},
		OpText{Text: " "},
	)

	// type
	var c Color
	var s string

	switch t {
	case Normal:
		c, s = White, "-"
	case Info:
		c, s = Green, "I"
	case Error:
		c, s = Red, "E"
	default:
		c, s = Yellow, "?"
	}

	p.out.Do(OpColor{Color: c},
		OpText{s},
		OpReset{},
		OpText{Text: " "},
	)

	// tag
	if tag != "" {
		p.out.Do(OpColor{Color: p.getTagColor(tag)},
			OpText{Text: fmt.Sprintf("[%s]", tag)},
			OpReset{},
			OpText{Text: " "},
		)
	}

	// text
	for _, op := range ops {
		p.out.Do(op)
	}

	// reset + newline
	p.out.Do(
		OpReset{},
		OpText{Text: "\n"},
	)
}

func (p *LinePrinter) getTagColor(tag string) Color {
	if c, ok := p.TagColors[tag]; ok {
		return c
	}
	return p.tagColor
}

func (p *LinePrinter) getLevelColor(level int) Color {
	return p.levelColor
}
