package term

import (
	"errors"
	"github.com/cryptopunkscc/astrald/streams"
	"regexp"
	"strconv"
	"strings"
)

var re = regexp.MustCompile(`%[+-]?\d*\.?\d*[svdf]`)

func Format(f string, v ...any) (a []any) {
	for {
		loc := re.FindIndex([]byte(f))
		if loc == nil {
			a = append(a, f)
			return
		}

		if loc[0] > 0 {
			a = append(a, f[:loc[0]])
		}

		format := f[loc[0]:loc[1]]

		var align *aligner

		if len(format) > 2 {
			w, err := strconv.Atoi(format[1 : len(format)-1])
			if err == nil {
				align = &aligner{
					w: w,
				}
			}
		}

		if len(v) > 0 {
			if align != nil {
				align.v = v[0]
				v[0] = align
			}
			a = append(a, v[0])
			v = v[1:]
		} else {
			a = append(a, errors.New("{!MISSING}"))
		}

		f = f[loc[1]:]
	}
}

func Format2(f string, v ...any) (a []any) {
	for len(f) > 0 {
		idx := strings.IndexByte(f, '%')
		if idx == -1 || idx == len(f)-1 {
			a = append(a, f)
			return
		}

		switch f[idx+1] {
		case 'v', 's':
			a = append(a, f[:idx])
			f = f[idx+2:]

		default:
			a = append(a, f[:idx+1])
			f = f[idx+1:]
			continue
		}

		a = append(a, &aligner{v: v[0], w: -26})
		v = v[1:]
	}
	return
}

var _ printerTo = &aligner{}

type aligner struct {
	v    any
	w    int
	pre  string
	post string
}

func (a aligner) PrintTo(printer *Printer, renderer Renderer) {
	wcount := streams.NewWriteCounter(nil)
	printer.print(NewBasicTerminal(wcount), a.v)

	l := int(wcount.Total())

	var r bool
	var m = a.w
	if m < 0 {
		m = m * -1
		r = true
	}
	var p int
	if l <= m {
		p = m - l
	}
	rnd := &cutRenderer{Renderer: renderer}
	rnd.Limit = m
	if (l > m) && r {
		rnd.Skip = l - m
	}
	if r {
		if len(a.pre) == 0 {
			renderer.Text(strings.Repeat(" ", p))
		} else {
			renderer.Text(strings.Repeat(a.pre, p))
		}
	}
	printer.PrintTo(rnd, a.v)
	if !r {
		if len(a.post) == 0 {
			renderer.Text(strings.Repeat(" ", p))
		} else {
			renderer.Text(strings.Repeat(a.post, p))
		}
	}
}

var _ Renderer = &cutRenderer{}

type cutRenderer struct {
	Renderer Renderer
	Skip     int
	Limit    int
}

func (c *cutRenderer) SetColor(color Color) {
	c.Renderer.SetColor(color)
}

func (c *cutRenderer) Text(s string) {
	if c.Skip == 0 {
		c.limitText(s)
		return
	}

	var m = min(c.Skip, len(s))
	c.Skip -= m
	s = s[m:]
	if len(s) >= 0 {
		c.limitText(s)
	}
}

func (c *cutRenderer) limitText(s string) {
	if c.Limit == 0 {
		return
	}

	var m = min(c.Limit, len(s))
	c.Limit -= m
	c.Renderer.Text(s[:m])
}
