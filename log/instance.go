package log

import (
	"os"
)

var instance = &Logger{
	out:       os.Stdout,
	emColor:   yellow,
	tagColor:  gray,
	timeColor: gray,
}

func Instance() *Logger {
	return instance
}

func Printf(format string, v ...interface{}) {
	instance.Log(format, v...)
}

func Log(format string, v ...interface{}) {
	instance.Log(format, v...)
}

func Logv(level int, format string, v ...interface{}) {
	instance.Logv(level, format, v...)
}

func Info(format string, v ...interface{}) {
	instance.Info(format, v...)
}

func Infov(level int, format string, v ...interface{}) {
	instance.Infov(level, format, v...)
}

func Error(format string, v ...interface{}) {
	instance.Error(format, v...)
}

func Errorv(level int, format string, v ...interface{}) {
	instance.Errorv(level, format, v...)
}

func Tag(tag string) *Logger {
	return instance.Tag(tag)
}
