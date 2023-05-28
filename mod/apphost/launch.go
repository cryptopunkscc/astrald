package apphost

import (
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/apphost/proto"
	"os/exec"
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

func (mod *Module) Launch(appName string, args []string, env []string) error {
	app, found := mod.config.Apps[appName]
	if !found {
		return errors.New("app not found")
	}

	identity, err := mod.node.Resolver().Resolve(app.Identity)
	if err != nil {
		return err
	}

	return mod.LaunchRuntime(app.Runtime, app.Path, identity, args, env)
}

func (mod *Module) LaunchRuntime(runtime string, path string, identity id.Identity, args []string, env []string) error {
	rbin, found := mod.config.Runtime[runtime]
	if !found {
		return errors.New("unsupported runtime")
	}

	return mod.LaunchRaw(rbin, identity, append([]string{path}, args...), env)
}

func (mod *Module) LaunchRaw(path string, identity id.Identity, args []string, env []string) error {
	var token = mod.createToken(identity)
	var log = mod.log.Tag(mod.node.Resolver().DisplayName(identity))

	cmd := exec.Command(path, args...)
	cmd.Env = env
	cmd.Env = append(cmd.Env, proto.EnvKeyAddr+"="+mod.getListeners())
	cmd.Env = append(cmd.Env, proto.EnvKeyToken+"="+token)
	cmd.Stdout = LogWriter{Log: log.Logv}
	cmd.Stderr = LogWriter{Log: log.Errorv}

	return cmd.Run()
}
