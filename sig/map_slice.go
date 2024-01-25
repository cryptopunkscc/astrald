package sig

func MapSlice[A, B any](in []A, conv func(A) (B, error)) ([]B, error) {
	var out []B
	for _, a := range in {
		b, err := conv(a)
		if err != nil {
			return nil, err
		}
		out = append(out, b)
	}

	return out, nil
}
