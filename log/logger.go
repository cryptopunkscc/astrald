package log

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"time"
)

type Type int

const (
	Normal = Type(iota)
	Info
	Error
)

const fullTimestamp = "2006-01-02 15:04:05.000"
const shortTimestamp = "15:04:05.000"

type FormatFunc func(v any) ([]Op, bool)

type Logger struct {
	parent     *Logger
	mu         sync.Mutex
	printer    Printer
	tag        string
	formatters []FormatFunc
	nestedTag  bool
}

func NewLogger(printer Printer) *Logger {
	return &Logger{
		printer:    printer,
		formatters: []FormatFunc{},
	}
}

func (l *Logger) Log(format string, v ...interface{}) {
	l.Logf(Normal, 0, time.Now(), l.getTag(), format, v...)
}

func (l *Logger) Logv(level int, format string, v ...interface{}) {
	l.Logf(Normal, level, time.Now(), l.getTag(), format, v...)
}

func (l *Logger) Info(format string, v ...interface{}) {
	l.Logf(Info, 0, time.Now(), l.getTag(), format, v...)
}

func (l *Logger) Infov(level int, format string, v ...interface{}) {
	l.Logf(Info, level, time.Now(), l.getTag(), format, v...)
}

func (l *Logger) Error(format string, v ...interface{}) {
	l.Logf(Error, 0, time.Now(), l.getTag(), format, v...)
}

func (l *Logger) Errorv(level int, format string, v ...interface{}) {
	l.Logf(Error, level, time.Now(), l.getTag(), format, v...)
}

func (l *Logger) Logf(t Type, level int, ts time.Time, tag string, f string, v ...interface{}) {
	l.printer.Log(t, level, ts, tag, l.toOps(f, v...)...)
}

func (l *Logger) SetNestedTag(nestedTag bool) {
	l.nestedTag = nestedTag
}

func (l *Logger) Sprintf(f string, v ...any) string {
	var buf = &bytes.Buffer{}
	var out = NewMonoOutput(buf)

	ops := l.toOps(f, v...)
	for _, op := range ops {
		out.Do(op)
	}

	return buf.String()
}

func (l *Logger) PushFormatFunc(fn FormatFunc) {
	l.formatters = append(l.formatters, fn)
}

func (l *Logger) Tag(tag string) *Logger {
	tagged := NewLogger(l.printer)
	tagged.parent = l
	tagged.tag = tag
	return tagged
}

func (l *Logger) Root() *Logger {
	if l.parent != nil {
		return l.parent.Root()
	}
	return l
}

func (l *Logger) Format(v any) ([]Op, bool) {
	for _, fn := range l.formatters {
		if ops, ok := fn(v); ok {
			return ops, true
		}
	}

	if l.parent != nil {
		return l.parent.Format(v)
	}

	return nil, false
}

func (l *Logger) toOps(f string, v ...any) []Op {
	var ops = make([]Op, 0)

	for {
		if len(f) == 0 {
			break
		}

		idx := strings.IndexByte(f, '%')
		if idx == -1 {
			ops = append(ops, OpText{Text: f})
			break
		}

		if idx > 0 {
			ops = append(ops, OpText{Text: f[:idx]})
			f = f[idx:]
		}

		f = f[1:]
		if len(f) == 0 {
			ops = append(ops, OpText{Text: "%"})
			break
		}

		switch f[0] {
		case 'v', 's', 'd', 'f', 't':
			f = f[1:]

			if len(v) == 0 {
				ops = append(ops,
					OpColor{Color: Red},
					OpText{Text: "!{MISSING_ARG}"},
					OpReset{},
				)
				continue
			}

			nv := v[0]
			v = v[1:]

			if vops, ok := l.Format(nv); ok {
				ops = append(ops, vops...)
				ops = append(ops, OpReset{})
				continue
			}

			ops = append(ops, OpText{Text: fmt.Sprintf("%v", nv)})

		default:
			ops = append(ops, OpText{Text: "%"})
			continue
		}
	}

	return ops
}

func (l *Logger) getTag() string {
	var prefix string
	if l.nestedTag && l.parent != nil {
		pt := l.parent.getTag()
		if len(pt) > 0 {
			prefix = pt + "/"
		}
	}
	return prefix + l.tag
}
