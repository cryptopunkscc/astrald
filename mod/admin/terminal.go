package admin

import (
	"bufio"
	"fmt"
	"io"
)

type Terminal struct {
	io.ReadWriter
}

func NewTerminal(rw io.ReadWriter) *Terminal {
	return &Terminal{ReadWriter: rw}
}

func (t *Terminal) Printf(f string, v ...any) {
	fmt.Fprintf(t, f, v...)
}

func (t *Terminal) Println(v ...any) {
	fmt.Fprintln(t, v...)
}

func (t *Terminal) Scanf(f string, v ...any) {
	fmt.Fscanf(t, f, v...)
}

func (t *Terminal) ScanLine() (string, error) {
	var scanner = bufio.NewScanner(t)
	if !scanner.Scan() {
		return "", scanner.Err()
	}
	return scanner.Text(), nil
}
