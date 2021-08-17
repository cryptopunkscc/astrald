package internal

import (
	_bytes "bytes"
	"os/exec"
	"strings"
)

func GetMimeType(bytes []byte) (string, error) {
	cmd := exec.Command("file", "--mime-type", "-")
	cmd.Stdin = _bytes.NewBuffer(bytes)

	res, err := cmd.Output()
	if err != nil {
		return "", err
	}

	sep := strings.SplitAfterN(string(res), ":", 2)
	typ := strings.TrimSpace(sep[1])

	return typ, nil
}
