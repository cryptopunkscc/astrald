package net

import "context"

func Advertise(ctx context.Context) error {
	for _, drv := range drivers {
		drv.Advertise(ctx)
	}
	return nil
}
