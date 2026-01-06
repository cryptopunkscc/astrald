package channel

import (
	"errors"
	"fmt"
	"strings"
)

type ParsedText struct {
	Type string // the $type part
	Enc  string // " " (space) for plain, ":" for base64
	Text string // the remaining text
}

func ParseText(input string) (ParsedText, error) {
	var result ParsedText

	// Input must start with "#["
	if len(input) < 3 || !strings.HasPrefix(input, "#[") {
		return result, errors.New("input must start with '#['")
	}

	// Find the closing bracket ']'
	bracketIdx := strings.Index(input[2:], "]")
	if bracketIdx == -1 {
		return result, errors.New("missing closing ']' after type")
	}
	bracketIdx += 2 // Adjust for the starting "#[" offset

	typePart := input[2:bracketIdx]
	if strings.ContainsAny(typePart, " \t\n\r") {
		return result, errors.New("type cannot contain whitespace")
	}

	result.Type = typePart

	// After the closing ']', we expect either ' ' or ':'
	if bracketIdx >= len(input) {
		return result, errors.New("unexpected end of input after type")
	}

	encChar := input[bracketIdx+1]
	if encChar != ' ' && encChar != ':' {
		return result, fmt.Errorf("expected ' ' or ':' after type, got '%c'", encChar)
	}

	result.Enc = string(encChar)

	// The rest is the text
	if bracketIdx+2 <= len(input) {
		result.Text = input[bracketIdx+2:]
	} else {
		result.Text = ""
	}

	return result, nil
}
