package channel

import (
	"errors"
	"strings"
)

type ParsedText struct {
	Type string // the $type part
	Enc  string // "text" or "base64" or "unknown" or "none"
	Text string // the remaining text
}

// ParseText splits a channel line into its #[type] header, encoding, and remaining text.
// The encoding is inferred from the first byte after the header: space/tab is "text",
// "=" or ":" is "base64", a bare header is "none", anything else is "unknown".
func ParseText(line string) (ParsedText, error) {
	var parsed = ParsedText{Enc: "none"}

	// parse the type header
	if !strings.HasPrefix(line, "#[") {
		return ParsedText{}, errors.New("type header missing")
	}

	typeEndIdx := strings.Index(line, "]")
	if typeEndIdx == -1 {
		return ParsedText{}, errors.New("type header missing")
	}
	parsed.Type, line = line[2:typeEndIdx], line[typeEndIdx+1:]

	// if the line is just the type, we're done
	if len(line) == 0 {
		return parsed, nil
	}

	// parse the encoding
	parsed.Enc, parsed.Text = line[0:1], line[1:]
	switch parsed.Enc {
	case " ", "\t":
		parsed.Enc = "text"
	case "=", ":":
		parsed.Enc = "base64"
	default:
		parsed.Enc = "unknown"
	}

	return parsed, nil
}
