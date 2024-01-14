package setup

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Dialogue struct {
	io.ReadWriter
	scanner *bufio.Scanner
}

func NewDialogue(rw io.ReadWriter) *Dialogue {
	return &Dialogue{
		ReadWriter: rw,
		scanner:    bufio.NewScanner(rw),
	}
}

func (d *Dialogue) Say(f string, v ...any) {
	if !strings.HasSuffix(f, "\n") {
		f = f + "\n"
	}
	fmt.Fprintf(d, f, v...)
}

func (d *Dialogue) Ask(f string, v ...any) (string, error) {
	if !strings.HasSuffix(f, "\n") && !strings.HasSuffix(f, " ") {
		f = f + " "
	}
	fmt.Fprintf(d, f, v...)
	return d.ReadLine()
}

func (d *Dialogue) AskInt(f string, v ...any) (int, error) {
	for {
		s, err := d.Ask(f, v...)
		if err != nil {
			return 0, err
		}

		i, err := strconv.Atoi(s)
		if err != nil {
			d.Say("Please enter a number.")
			continue
		}
		return i, nil
	}
}

func (d *Dialogue) AskBool(f string, v ...any) (bool, error) {
	for {
		s, err := d.Ask(f, v...)
		if err != nil {
			return false, err
		}

		switch strings.ToLower(s) {
		case "y", "yes", "true":
			return true, nil

		case "n", "no", "false":
			return false, nil
		}
		d.Say("Please answer yes/no.")
	}
}

func (d *Dialogue) ReadLine() (string, error) {
	if d.scanner.Scan() {
		return d.scanner.Text(), nil
	}

	return "", io.EOF
}
