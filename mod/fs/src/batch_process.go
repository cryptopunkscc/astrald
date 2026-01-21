package fs

// batchProcess processes items from a yield-style iterator in batches.
func batchProcess[T any](
	iter func(yield func(T) error) error,
	process func(batch []T) error,
	batchSize int,
) error {
	batch := make([]T, 0, batchSize)

	err := iter(func(item T) error {
		batch = append(batch, item)
		if len(batch) >= batchSize {
			if err := process(batch); err != nil {
				return err
			}
			batch = batch[:0]
		}
		return nil
	})
	if err != nil {
		return err
	}

	if len(batch) > 0 {
		return process(batch)
	}

	return nil
}
