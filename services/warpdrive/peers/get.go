package peers

import (
	"encoding/json"
	"github.com/cryptopunkscc/astrald/services"
	"os"
	"path/filepath"
)

const peerInfoFilename = "peers"

type cacheAddr struct {
	Network string
	Address string
}

type cache map[string][]cacheAddr

func Get() ([]string, error) {
	// Reading peers json from file
	bytes, err := os.ReadFile(filepath.Join(services.AstralHome, peerInfoFilename))
	if err != nil {
		return nil, err
	}

	// Deserializing json
	c := make(cache)
	err = json.Unmarshal(bytes, &c)
	if err != nil {
		return nil, err
	}

	// Getting ids
	i := 0
	p := make([]string, 0, len(c))
	for id := range c {
		p[i] = id
		i++
	}
	return p, err
}
