//go:build windows

package fs

import "errors"

func DiskUsage(path string) (usage *DiskUsageInfo, err error) {
	return nil, errors.ErrUnsupported // TODO
}
