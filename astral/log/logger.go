package log

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/term"
	"sync"
	"time"
)

type Logger struct {
	Level  int
	tag    Tag
	id     *astral.Identity
	p      term.Printer
	mu     sync.Mutex
	parent *Logger
}

func NewLogger(p term.Printer, id *astral.Identity, tag Tag) *Logger {
	return &Logger{tag: tag, id: id, p: p}
}

func (l *Logger) Logf(t Type, level Level, ts time.Time, tag Tag, f string, v ...interface{}) {
	l.lock()
	defer l.unlock()

	l.log(NewEntry(l.id, level, t, tag, f, v...))
}

func (l *Logger) log(e *Entry) {
	if e.Level > Level(l.Level) {
		return
	}

	term.Printf(l.p, "%v\n", e)
}

func (l *Logger) Log(format string, v ...interface{}) {
	l.Logf(Normal, 0, time.Now(), l.tag, format, v...)
}

func (l *Logger) Logv(level int, format string, v ...interface{}) {
	l.Logf(Normal, Level(level), time.Now(), l.tag, format, v...)
}

func (l *Logger) Info(format string, v ...interface{}) {
	l.Logf(Info, 0, time.Now(), l.tag, format, v...)
}

func (l *Logger) Infov(level int, format string, v ...interface{}) {
	l.Logf(Info, Level(level), time.Now(), l.tag, format, v...)
}

func (l *Logger) Error(format string, v ...interface{}) {
	l.Logf(Error, 0, time.Now(), l.tag, format, v...)
}

func (l *Logger) Errorv(level int, format string, v ...interface{}) {
	l.Logf(Error, Level(level), time.Now(), l.tag, format, v...)
}

func (l *Logger) lock() {
	l.mu.Lock()
	if l.parent != nil {
		l.parent.lock()
	}
}

func (l *Logger) unlock() {
	if l.parent != nil {
		l.parent.unlock()
	}
	l.mu.Unlock()
}

func (l *Logger) Tag(tag Tag) *Logger {
	return &Logger{
		parent: l,
		tag:    tag,
		id:     l.id,
		p:      l.p,
		Level:  l.Level,
	}
}
