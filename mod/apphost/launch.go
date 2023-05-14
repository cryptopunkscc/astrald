package apphost

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/cryptopunkscc/astrald/mod/apphost/proto"
	"os/exec"
)

func (mod *Module) Launch(runtime string, path string) error {
	rbin, found := mod.config.Runtime[runtime]
	if !found {
		return errors.New("unsupported runtime")
	}

	sum := sha256.New()
	sum.Write([]byte(path))
	token := hex.EncodeToString(sum.Sum(nil))
	mod.tokens[token] = path

	cmd := exec.Command(rbin, path)
	cmd.Env = append(cmd.Env, proto.EnvKeyAddr+"="+mod.getListeners())
	cmd.Env = append(cmd.Env, proto.EnvKeyToken+"="+token)

	return cmd.Run()
}
