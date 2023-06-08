package log

import "io"

var _ Output = &ColorOutput{}

type ColorOutput struct {
	io.Writer
}

func (out *ColorOutput) Do(ops ...Op) {
	for _, op := range ops {
		switch op := op.(type) {
		case OpText:
			out.Write([]byte(op.Text))

		case OpReset:
			out.Write([]byte("\033[0m"))

		case OpColor:
			out.setColor(op.Color)

		case OpBackgroundColor:
			out.setBackgroundColor(op.Color)

		case OpBold:
			out.setBold(op.Bold)

		case OpFaint:
			out.setFaint(op.Faint)

		case OpItalic:
			out.setItalic(op.Italic)

		case OpUnderline:
			out.setUnderline(op.Underline)

		case OpBlink:
			out.setBlink(op.Blink)

		case OpStrike:
			out.setStrike(op.Strike)
		}
	}
}

func NewColorOutput(writer io.Writer) *ColorOutput {
	return &ColorOutput{Writer: writer}
}

func (out *ColorOutput) setColor(color Color) {
	var s string
	switch color {
	case Black:
		s = "\033[30m"
	case Red:
		s = "\033[31m"
	case Green:
		s = "\033[32m"
	case Yellow:
		s = "\033[33m"
	case Blue:
		s = "\033[34m"
	case Magenta:
		s = "\033[35m"
	case Cyan:
		s = "\033[36m"
	case White:
		s = "\033[37m"

	case BrightBlack:
		s = "\033[90m"
	case BrightRed:
		s = "\033[91m"
	case BrightGreen:
		s = "\033[92m"
	case BrightYellow:
		s = "\033[93m"
	case BrightBlue:
		s = "\033[94m"
	case BrightMagenta:
		s = "\033[95m"
	case BrightCyan:
		s = "\033[96m"
	case BrightWhite:
		s = "\033[97m"
	}

	out.Write([]byte(s))
}

func (out *ColorOutput) setBackgroundColor(color Color) {
	var s string

	switch color {
	case Black:
		s = "\033[40m"
	case Red:
		s = "\033[41m"
	case Green:
		s = "\033[42m"
	case Yellow:
		s = "\033[43m"
	case Blue:
		s = "\033[44m"
	case Magenta:
		s = "\033[45m"
	case Cyan:
		s = "\033[46m"
	case White:
		s = "\033[47m"

	case BrightBlack:
		s = "\033[100m"
	case BrightRed:
		s = "\033[101m"
	case BrightGreen:
		s = "\033[102m"
	case BrightYellow:
		s = "\033[103m"
	case BrightBlue:
		s = "\033[104m"
	case BrightMagenta:
		s = "\033[105m"
	case BrightCyan:
		s = "\033[106m"
	case BrightWhite:
		s = "\033[107m"
	}

	out.Write([]byte(s))
}

func (out *ColorOutput) setBold(b bool) {
	if b {
		out.Write([]byte("\033[1m"))
	} else {
		out.Write([]byte("\033[22m"))
	}
}

func (out *ColorOutput) setFaint(b bool) {
	if b {
		out.Write([]byte("\033[2m"))
	} else {
		out.Write([]byte("\033[22m"))
	}
}

func (out *ColorOutput) setItalic(b bool) {
	if b {
		out.Write([]byte("\033[3m"))
	} else {
		out.Write([]byte("\033[23m"))
	}
}

func (out *ColorOutput) setUnderline(b bool) {
	if b {
		out.Write([]byte("\033[4m"))
	} else {
		out.Write([]byte("\033[24m"))
	}
}

func (out *ColorOutput) setBlink(b bool) {
	if b {
		out.Write([]byte("\033[5m"))
	} else {
		out.Write([]byte("\033[25m"))
	}
}

func (out *ColorOutput) setStrike(b bool) {
	if b {
		out.Write([]byte("\033[9m"))
	} else {
		out.Write([]byte("\033[29m"))
	}
}
