package log

type Output interface {
	LogEntry(*Entry)
}
