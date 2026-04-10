package log

import (
	"io"
	"sync"
)

// Printer is a log output that prints log entries to a writer using a Viewer for rendering.
type Printer struct {
	Viewer *Viewer
	Filter Filter
	w      io.Writer
	mu     sync.Mutex
}

func NewPrinter(w io.Writer) *Printer {
	return &Printer{w: w, Viewer: DefaultViewer}
}

func (output *Printer) LogEntry(entry *Entry) {
	// apply filter
	if output.Filter != nil && !output.Filter(entry) {
		return
	}

	str := output.Viewer.Render(Format("%v\n", entry)...)

	output.mu.Lock()
	defer output.mu.Unlock()

	_, _ = output.w.Write([]byte(str))
}
