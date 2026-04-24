package log

import (
	"io"
	"os"
	"strings"
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/fmt"
	"github.com/cryptopunkscc/astrald/sig"
)

type Logger struct {
	id      *astral.Identity
	w       io.Writer
	mu      sync.Mutex
	parent  *Logger
	prefix  []astral.Object
	filter  func(*Entry) bool
	loggers sig.Set[EntryLogger]
}

func (l *Logger) SetFilter(filter func(*Entry) bool) {
	if l.parent != nil {
		l.parent.SetFilter(filter)
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	l.filter = filter
}

type EntryLogger interface {
	LogEntry(*Entry)
}

func New(id *astral.Identity) *Logger {
	return &Logger{
		id: id,
		w:  os.Stdout,
	}
}

func (l *Logger) Log(format string, v ...interface{}) {
	l.logf(0, format, v...)
}

func (l *Logger) Logv(level int, format string, v ...interface{}) {
	l.logf(uint8(level), format, v...)
}

func (l *Logger) Info(format string, v ...interface{}) {
	l.logf(0, format, v...)
}

func (l *Logger) Infov(level int, format string, v ...interface{}) {
	l.logf(uint8(level), format, v...)
}

func (l *Logger) Error(format string, v ...interface{}) {
	l.logf(0, format, v...)
}

func (l *Logger) Errorv(level int, format string, v ...interface{}) {
	l.logf(uint8(level), format, v...)
}

func (l *Logger) SetPrefix(obj ...astral.Object) *Logger {
	return &Logger{
		parent: l,
		id:     l.id,
		prefix: obj,
	}
}

func (l *Logger) Tag(tag Tag) *Logger {
	return l.SetPrefix(&tag)
}

func (l *Logger) AppendTag(tag Tag) *Logger {
	return l.SetPrefix(append(l.prefix, &tag)...)
}

func (l *Logger) AddLogger(el EntryLogger) {
	l.root().loggers.Add(el)
}

func (l *Logger) RemoveLogger(el EntryLogger) {
	l.root().loggers.Remove(el)
}

func (l *Logger) logf(level uint8, f string, v ...interface{}) {
	var items = l.prefix

	f = strings.ReplaceAll(f, "\n", "\\\\n")

	for _, a := range fmt.Format(f, v...) {
		obj := astral.Adapt(a)
		if obj == nil {
			if view, ok := a.(fmt.View); ok {
				obj = astral.NewString32(view.Render())
			} else {
				obj = astral.NewString32(astral.Stringify(a))
			}
		}
		items = append(items, obj)
	}

	entry := NewEntry(l.id, level, items...)

	l.root().logEntry(entry)
}

func (l *Logger) logEntry(e *Entry) {
	l.mu.Lock()
	defer l.mu.Unlock()

	f := l.root().filter
	if f == nil || f(e) {
		fmt.Fprintf(l.w, "%v\n", e)
	}

	for _, l := range l.root().loggers.Clone() {
		l.LogEntry(e)
	}
}

func (l *Logger) root() *Logger {
	if l.parent != nil {
		return l.parent.root()
	}
	return l
}
