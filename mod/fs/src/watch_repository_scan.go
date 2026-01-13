package fs

import (
	"context"
	"errors"
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

		err := repo.rescan(ctx)
		if err == nil {
			done <- nil
			return
		}

		// Normalize cancellation.
		if errors.Is(err, context.Canceled) {
			done <- context.Canceled
			return
		}

		done <- err
	}()

	return ScanHandle{Cancel: cancel, Done: done}
}

var errScanCanceled = errors.New("scan canceled")

func (repo *WatchRepository) rescan(ctx *astral.Context) error {
	err := filepath.WalkDir(repo.root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Check context early for prompt cancellation.
		select {
		case <-ctx.Done():
			return errScanCanceled
		default:
		}

		// Check if the entry is a regular file
		if !entry.Type().IsRegular() {
			return nil
		}

		err = repo.mod.checkIndexEntry(path)
		if err != nil {
			repo.mod.pathIndexer.MarkDirtyOwned(repo.label, path)
		}

		return nil
	})
	if err != nil {
		if errors.Is(err, errScanCanceled) {
			return ctx.Err()
		}
		return err
	}

	return nil
}
