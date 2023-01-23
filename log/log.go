package log

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"time"
)

type Logger struct {
	parent    *Logger
	emColor   string
	tagColor  string
	timeColor string
	tag       string
	out       io.Writer
	mu        sync.Mutex
}

const (
	TypeNormal = iota
	TypeInfo
	TypeError
)

const (
	LevelNormal = iota
	LevelDebug
)

const fullTimestamp = "2006-01-02 15:04:05.000"
const shortTimestamp = "15:04:05.000"

var HideDate bool

func Sformat(format string, v ...interface{}) string {
	return instance.Sformat(format, v...)
}

func SetTimeColor(color string) {
	instance.SetTimeColor(color)
}

func EmColor() string {
	return instance.EmColor()
}

func SetEmColor(color string) {
	instance.SetEmColor(color)
}

func SetTagColor(color string) {
	instance.SetTagColor(color)
}

func (l *Logger) Sformat(format string, v ...interface{}) string {
	var buf = &bytes.Buffer{}
	l.Capture(buf).format(format, v...)
	return string(buf.Bytes())
}

func (l *Logger) Log(format string, v ...interface{}) {
	l.log(TypeNormal, 0, format, v...)
}

func (l *Logger) Logv(level int, format string, v ...interface{}) {
	l.log(TypeNormal, level, format, v...)
}

func (l *Logger) Info(format string, v ...interface{}) {
	l.log(TypeInfo, 0, format, v...)
}

func (l *Logger) Infov(level int, format string, v ...interface{}) {
	l.log(TypeInfo, level, format, v...)
}

func (l *Logger) Error(format string, v ...interface{}) {
	l.log(TypeError, 0, format, v...)
}

func (l *Logger) Errorv(level int, format string, v ...interface{}) {
	l.log(TypeError, level, format, v...)
}

func (l *Logger) SetTimeColor(color string) {
	l.root().timeColor = color
}

func (l *Logger) SetEmColor(color string) {
	l.root().emColor = color
}

func (l *Logger) EmColor() string {
	return l.root().emColor
}

func (l *Logger) SetTagColor(color string) {
	l.root().tagColor = color
}

func (l *Logger) Tag(tag string) (s *Logger) {
	s = l.sub()
	s.tag = tag
	return
}

func (l *Logger) Capture(w io.Writer) (s *Logger) {
	s = l.sub()
	s.out = w
	s.parent = nil
	return
}

func (l *Logger) sub() *Logger {
	return &Logger{
		parent: l,
		tag:    l.tag,
		out:    l.out,
	}
}

func (l *Logger) format(format string, v ...interface{}) {
	fmt.Fprintf(l.out, format, formatArgs(v)...)
}

func (l *Logger) getEmColor() string {
	return l.root().emColor
}

func (l *Logger) getTimeColor() string {
	return l.root().timeColor
}

func (l *Logger) getTagColor() string {
	return l.root().tagColor
}

func (l *Logger) formatTag() string {
	if l.tag == "" {
		return ""
	}

	return l.getTagColor() + "[" + l.tag + "] " + reset
}

func (l *Logger) root() *Logger {
	r := l
	for r.parent != nil {
		r = r.parent
	}
	return r
}

func (l *Logger) log(kind int, level int, format string, v ...interface{}) {
	if getTagLevel(l.tag) < level {
		return
	}

	var root = l.root()

	root.mu.Lock()
	defer root.mu.Unlock()

	if format[len(format)-1] != '\n' {
		format += "\n"
	}

	var timestamp string
	if HideDate {
		timestamp = time.Now().Format(shortTimestamp)
	} else {
		timestamp = time.Now().Format(fullTimestamp)
	}

	fmt.Fprintf(l.out, "%s%s%s ", l.getTimeColor(), timestamp, reset)

	switch kind {
	case TypeNormal:
		fmt.Fprintf(l.out, "%s-%s ", gray, reset)
	case TypeInfo:
		fmt.Fprintf(l.out, "%sI%s ", green, reset)
	case TypeError:
		fmt.Fprintf(l.out, "%sE%s ", red, reset)
	}

	fmt.Fprintf(l.out, l.formatTag()+format, formatArgs(v)...)
}
