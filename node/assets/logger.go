package assets

type Logger interface {
	Errorv(level int, format string, v ...interface{})
}
