package term

import "io"

const defaultTextColor = White

var _ Renderer = &LinuxTerminal{}

type LinuxTerminal struct {
	output           io.Writer
	defaultTextColor Color
}

func NewLinuxTerminal(output io.Writer) *LinuxTerminal {
	c := &LinuxTerminal{
		output:           output,
		defaultTextColor: defaultTextColor,
	}
	c.SetColor(c.defaultTextColor)
	return c
}

func (p *LinuxTerminal) Text(text string) {
	p.output.Write([]byte(text))
}

func (p *LinuxTerminal) SetColor(color Color) {
	var s string
	switch color {
	case Reset:
		if p.defaultTextColor != Reset {
			p.SetColor(p.defaultTextColor)
		}
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

	p.output.Write([]byte(s))
}
