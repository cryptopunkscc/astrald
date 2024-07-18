package dir

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"net"
	"strings"
)

var _ dir.Resolver = &DNS{}

type DNS struct {
	*Module
}

func (dns DNS) Resolve(s string) (identity id.Identity, err error) {
	if strings.Contains(s, ".") {
		domain := "_astral." + s
		txtRecords, err := net.LookupTXT(domain)
		if err != nil {
			dns.log.Errorv(1, "Error looking up TXT records for %v: %v\n", domain, err)
			return identity, err
		}

		for _, record := range txtRecords {
			if !strings.HasPrefix(record, "id=") {
				continue
			}

			identity, err = id.ParsePublicKeyHex(record[3:])
			if err == nil {
				return identity, nil
			}
		}
	}

	return id.Identity{}, fmt.Errorf("cannot resolve")
}

func (dns DNS) DisplayName(identity id.Identity) string {
	return ""
}
