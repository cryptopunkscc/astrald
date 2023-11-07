//go:build sqlite_native

package assets

import "gorm.io/driver/sqlite"

func init() {
	DBOpener = sqlite.Open
}
