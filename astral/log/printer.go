package log

import (
	"fmt"
	"io"
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
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

	var line = Format("[%v] (%v) %v ",
		entry.Origin,
		&entry.Level,
		&entry.Time,
	)

	line = append(line, entry.Objects...)
	line = append(line, (*astral.String8)(astral.NewString("\n")))

	str := output.Viewer.Render(line...)

	output.mu.Lock()
	defer output.mu.Unlock()

	_, _ = fmt.Fprintf(output.w, str)
}
