package dir

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"net"
	"regexp"
	"strings"
)

var domainRegex = regexp.MustCompile(`^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]$`)
var _ dir.Resolver = &DNS{}

type DNS struct {
	*Module
}

func isValidDomain(domain string) bool {
	return domainRegex.MatchString(domain)
}

func (dns DNS) ResolveIdentity(s string) (identity *astral.Identity, err error) {
	if !isValidDomain(s) {
		return &astral.Identity{}, fmt.Errorf("cannot resolve")
	}

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

		identity, err = astral.ParseIdentity(record[3:])
		if err == nil {
			return identity, nil
		}
	}

	return &astral.Identity{}, fmt.Errorf("cannot resolve")
}

func (dns DNS) DisplayName(identity *astral.Identity) string {
	return ""
}
