package admin

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"io"
	"strings"
	"time"
)

const useColorTerminal = true

var _ admin.Terminal = &ColorTerminal{}

type ColorTerminal struct {
	userIdentity *astral.Identity
	color        bool
	log          *log.Logger
	output       log.Output
	io.ReadWriter
}

func NewColorTerminal(rw astral.Conn, logger *log.Logger) *ColorTerminal {
	l := logger.Tag("")
	l.PushFormatFunc(func(v any) ([]log.Op, bool) {
		s, ok := v.(string)
		if !ok {
			return nil, false
		}
		return []log.Op{log.OpText{Text: s}}, true
	})

	l.PushFormatFunc(func(v any) ([]log.Op, bool) {
		s, ok := v.(admin.Header)
		if !ok {
			return nil, false
		}
		return []log.Op{
			log.OpColor{Color: log.White},
			log.OpText{Text: strings.ToUpper(string(s))},
			log.OpReset{},
		}, true
	})

	l.PushFormatFunc(func(v any) ([]log.Op, bool) {
		s, ok := v.(admin.Keyword)
		if !ok {
			return nil, false
		}
		return []log.Op{
			log.OpColor{Color: log.Yellow},
			log.OpText{Text: string(s)},
			log.OpReset{},
		}, true
	})

	l.PushFormatFunc(func(v any) ([]log.Op, bool) {
		s, ok := v.(admin.Faded)
		if !ok {
			return nil, false
		}
		return []log.Op{
			log.OpColor{Color: log.BrightBlack},
			log.OpText{Text: string(s)},
			log.OpReset{},
		}, true
	})

	l.PushFormatFunc(func(v any) ([]log.Op, bool) {
		s, ok := v.(admin.Important)
		if !ok {
			return nil, false
		}
		return []log.Op{
			log.OpColor{Color: log.BrightRed},
			log.OpText{Text: string(s)},
			log.OpReset{},
		}, true
	})

	l.PushFormatFunc(func(v any) ([]log.Op, bool) {
		t, ok := v.(time.Time)
		if !ok {
			return nil, false
		}
		return []log.Op{
			log.OpText{Text: t.Format(timestampFormat)},
		}, true
	})
	return &ColorTerminal{
		color:        useColorTerminal,
		userIdentity: rw.RemoteIdentity(),
		ReadWriter:   rw,
		output:       log.NewColorOutput(rw),
		log:          l,
	}
}

func (t *ColorTerminal) Sprintf(f string, v ...any) string {
	var buf = &bytes.Buffer{}
	var out log.Output

	if t.color {
		out = log.NewColorOutput(buf)
	} else {
		out = log.NewMonoOutput(buf)
	}

	out.Do(t.log.Renderf(f, v...)...)

	return buf.String()
}

func (t *ColorTerminal) Printf(f string, v ...any) {
	if t.color {
		t.output.Do(t.log.Renderf(f, v...)...)
		return
	}
	fmt.Fprintf(t, f, v...)
}

func (t *ColorTerminal) Println(v ...any) {
	if t.color {
		for _, i := range v {
			ops, b := t.log.Render(i)
			if b {
				t.output.Do(ops...)
			} else {
				fmt.Fprintln(t, i)
			}
		}
		t.output.Do(log.OpText{Text: "\n"})
		return
	}
	fmt.Fprintln(t, v...)
}

func (t *ColorTerminal) Scanf(f string, v ...any) {
	fmt.Fscanf(t, f, v...)
}

func (t *ColorTerminal) ScanLine() (string, error) {
	var scanner = bufio.NewScanner(t)
	if !scanner.Scan() {
		return "", io.EOF
	}
	return scanner.Text(), nil
}

func (t *ColorTerminal) Color() bool {
	return t.color
}

func (t *ColorTerminal) SetColor(Color bool) {
	t.color = Color
}

func (t *ColorTerminal) UserIdentity() *astral.Identity {
	return t.userIdentity
}

func (t *ColorTerminal) SetUserIdentity(identity *astral.Identity) {
	t.userIdentity = identity
}
