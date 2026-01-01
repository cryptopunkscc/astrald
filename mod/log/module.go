package log

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

const ModuleName = "log"

type Module interface {
}

func init() {
	lipgloss.SetColorProfile(termenv.TrueColor)
}
