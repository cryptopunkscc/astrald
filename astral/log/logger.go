package log

import (
	"os"
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
)

type Logger struct {
	id     *astral.Identity
	mu     sync.Mutex
	parent *Logger
	prefix []astral.Object
	output Output
}

func New(id *astral.Identity, output Output) *Logger {
	if output == nil {
		output = NewPrinter(os.Stdout)
	}

	return &Logger{
		id:     id,
		output: output,
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

func (l *Logger) SetOutput(o Output) {
	l.root().setOutput(o)
}

func (l *Logger) Output() Output {
	return l.root().getOutput()
}

func (l *Logger) logf(level uint8, f string, v ...interface{}) {
	obj := append(l.prefix, Format(f, v...)...)

	entry := NewEntry(l.id, level, obj...)

	l.Output().LogEntry(entry)
}

func (l *Logger) setOutput(o Output) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.output = o
}

func (l *Logger) getOutput() Output {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.output
}

func (l *Logger) root() *Logger {
	if l.parent != nil {
		return l.parent.root()
	}
	return l
}
