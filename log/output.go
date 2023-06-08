package log

import (
	"io"
	"strings"
)

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

var _ io.WriterTo

type Op any

type OpText struct {
	Text string
}

type OpReset struct {
}

type OpColor struct {
	Color
}

type OpBackgroundColor struct {
	Color
}

type OpBold struct {
	Bold bool
}

type OpFaint struct {
	Faint bool
}

type OpItalic struct {
	Italic bool
}

type OpUnderline struct {
	Underline bool
}

type OpBlink struct {
	Blink bool
}

type OpStrike struct {
	Strike bool
}

type Output interface {
	Do(...Op)
}

func ParseColor(s string) Color {
	switch strings.ToLower(s) {
	case "black":
		return Black
	case "red":
		return Red
	case "green":
		return Green
	case "yellow":
		return Yellow
	case "blue":
		return Blue
	case "magenta":
		return Magenta
	case "cyan":
		return Cyan
	case "white":
		return White

	case "brightblack":
		return BrightBlack
	case "brightred":
		return BrightRed
	case "brightgreen":
		return BrightGreen
	case "brightyellow":
		return BrightYellow
	case "brightblue":
		return BrightBlue
	case "brightmagenta":
		return BrightMagenta
	case "brightcyan":
		return BrightCyan
	case "brightwhite":
		return BrightWhite
	}

	return Reset
}
