package fs

import (
	"os"
	"path/filepath"

	"github.com/cryptopunkscc/astrald/astral"
)

type ScanHandle struct {
	Cancel func()
	Done   <-chan error // receives exactly one error (nil on success) and then closes
}

// StartScan starts a background scan of repo.root that can be canceled via the returned handle.
//
// Done will always produce exactly one terminal result:
//   - nil on successful completion
//   - context.Canceled if canceled
//   - another error for failures
func (repo *WatchRepository) StartScan(parent *astral.Context) ScanHandle {
	// Cancel any previous scan.
	if repo.scan.Cancel != nil {
		repo.scan.Cancel()
	}

	ctx, cancel := parent.WithCancel()
	done := make(chan error, 1)

	go func() {
		defer close(done)
		done <- repo.rescan(ctx)
	}()

	return ScanHandle{Cancel: cancel, Done: done}
}

func (repo *WatchRepository) rescan(ctx *astral.Context) error {
	return filepath.WalkDir(repo.root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Check context early for prompt cancellation.
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Check if the entry is a regular file
		if !entry.Type().IsRegular() {
			return nil
		}

		err = repo.mod.checkIndexEntry(path)
		if err != nil {
			repo.mod.fileIndexer.AcquireRoot(path)
			repo.mod.fileIndexer.MarkDirty(path)
		}

		return nil
	})
}
