package log

import "os"

func NewFileOutput(file string) (Output, error) {
	f, err := os.Create(file)
	if err != nil {
		return nil, err
	}

	return NewMonoOutput(f), nil
}
