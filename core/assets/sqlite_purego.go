//go:build !sqlite_native

package assets

import "github.com/glebarez/sqlite"

func init() {
	dbOpen = sqlite.Open
}
