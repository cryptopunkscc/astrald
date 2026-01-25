package dir

import (
	"github.com/cryptopunkscc/astrald/astral"
)

const ModuleName = "dir"
const DBPrefix = "dir__"

type Module interface {
	SetAlias(*astral.Identity, string) error
	GetAlias(*astral.Identity) (string, error)
	ResolveIdentity(string) (*astral.Identity, error)
	DisplayName(*astral.Identity) string
	AddResolver(Resolver) error

	// SetFilter sets a function for a named filter
	SetFilter(name string, filter astral.IdentityFilter)

	// GetFilter returns the filter function for a named filter
	GetFilter(name string) astral.IdentityFilter

	// Filters method returns the list of registered filters
	Filters() []string

	// DefaultFilters returns the default list of filters applied to queries with no custom filters
	DefaultFilters() []string

	// SetDefaultFilters sets the default list of filters applied to queries with no custom filters
	SetDefaultFilters(filters ...string)

	// ApplyFilters returns true if any of the filters returns true for the provided identity
	ApplyFilters(identity *astral.Identity, filter ...string) bool
}

type Resolver interface {
	ResolveIdentity(string) (*astral.Identity, error)
	DisplayName(*astral.Identity) string
}
