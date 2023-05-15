package apphost

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/cryptopunkscc/astrald/mod/apphost/proto"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type LogWriter struct {
	lines string
	Log   func(level int, format string, v ...interface{})
}

func (w LogWriter) Write(p []byte) (n int, err error) {
	w.lines = w.lines + string(p)

	for {
		i := strings.Index(w.lines, "\n")
		if i == -1 {
			break
		}
		line := w.lines[:i]
		if w.Log != nil {
			w.Log(2, line)
		}
		w.lines = w.lines[i+1:]
	}

	return len(p), nil
}

func (mod *Module) Launch(runtime string, path string) error {
	rbin, found := mod.config.Runtime[runtime]
	if !found {
		return errors.New("unsupported runtime")
	}

	sum := sha256.New()
	sum.Write([]byte(path))
	token := hex.EncodeToString(sum.Sum(nil))
	mod.tokens[token] = path

	log := log.Tag(filepath.Base(path))

	cmd := exec.Command(rbin, path)
	cmd.Env = os.Environ() // TODO: rethink the security of this, maybe whitelist only some variables?
	cmd.Env = append(cmd.Env, proto.EnvKeyAddr+"="+mod.getListeners())
	cmd.Env = append(cmd.Env, proto.EnvKeyToken+"="+token)
	cmd.Stdout = LogWriter{Log: log.Logv}
	cmd.Stderr = LogWriter{Log: log.Errorv}

	return cmd.Run()
}
