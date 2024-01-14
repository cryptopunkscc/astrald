package policy

import "time"

type Intervals []int

var retryIvals = Intervals{5, 5, 5, 5, 5, 5, 15, 15, 15, 15, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 5 * 60, 5 * 60, 5 * 60, 5 * 60, 15 * 60}

func (ivals Intervals) At(i int) time.Duration {
	if i < len(ivals) {
		return time.Duration(ivals[i]) * time.Second
	} else {
		return time.Duration(ivals[len(ivals)-1]) * time.Second
	}
}
