package term

type Color int

const (
	Reset = Color(iota)
	Black
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
	BrightBlack
	BrightRed
	BrightGreen
	BrightYellow
	BrightBlue
	BrightMagenta
	BrightCyan
	BrightWhite
)

func C(c Color, v any) any {
	return &forceColor{
		v: v,
		c: c,
	}
}

var _ printerTo = &forceColor{}

type forceColor struct {
	v any
	c Color
}

func (f forceColor) PrintTo(p *Printer, r Renderer) {
	r.SetColor(f.c)
	p.PrintTo(stripColors{r: r}, f.v)
	r.SetColor(Reset)
}

type stripColors struct {
	r Renderer
}

func (stripColors) SetColor(color Color) {}

func (sc stripColors) Text(s string) {
	sc.r.Text(s)
}
