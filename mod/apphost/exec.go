package apphost

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/apphost/proto"
	"os/exec"
	"strconv"
	"strings"
)

func (mod *Module) Exec(identity id.Identity, path string, args []string, env []string) (*Exec, error) {
	var token = mod.createToken(identity)
	var log = mod.log.Tag(mod.node.Resolver().DisplayName(identity))

	e := &Exec{
		identity: identity,
		path:     path,
		args:     args,
		env:      env,
		token:    token,
		done:     make(chan struct{}),
		state:    "running",
	}

	cmd := exec.Command(path, args...)
	cmd.Env = env
	cmd.Env = append(cmd.Env, proto.EnvKeyAddr+"="+mod.getListeners())
	cmd.Env = append(cmd.Env, proto.EnvKeyToken+"="+token)
	cmd.Stdout = LogWriter{Log: log.Logv}
	cmd.Stderr = LogWriter{Log: log.Errorv}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	e.cmd = cmd

	go func() {
		e.err = cmd.Wait()
		if e.err != nil {
			if typed, ok := e.err.(*exec.ExitError); ok {
				code := typed.ExitCode()
				if code == -1 {
					e.state = "killed"
				} else {
					e.state = "exit " + strconv.Itoa(code)
				}
			} else {
				e.state = "error"
			}
		} else {
			e.state = "finished"
		}
		close(e.done)
	}()

	mod.execs = append(mod.execs, e)

	return e, nil
}

type Exec struct {
	identity id.Identity
	path     string
	args     []string
	env      []string
	token    string
	cmd      *exec.Cmd

	state string
	done  chan struct{}
	err   error
}

func (e *Exec) Kill() error {
	return e.cmd.Process.Kill()
}

func (e *Exec) State() string {
	return e.state
}

func (e *Exec) Identity() id.Identity {
	return e.identity
}

func (e *Exec) Path() string {
	return e.path
}

func (e *Exec) Args() []string {
	return e.args
}

func (e *Exec) Env() []string {
	return e.env
}

func (e *Exec) Token() string {
	return e.token
}

func (e *Exec) Err() error {
	return e.err
}

func (e *Exec) Done() <-chan struct{} {
	return e.done
}

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
