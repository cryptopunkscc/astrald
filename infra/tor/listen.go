package tor

import (
	"os"
	"path"
)

// TODO: figure out how to do driver cache nicely
func keyPath() string {
	home, _ := os.UserHomeDir()
	return path.Join(home, ".config", "astrald", "tor.key")
}
