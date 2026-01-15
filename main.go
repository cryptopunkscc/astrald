package main

import (
	"fmt"

	"github.com/cryptopunkscc/astrald/lib/astrald"
)

func main() {
	ctx := astrald.NewContext()

	ch, err := astrald.Services().Discover(ctx, true)
	if err != nil {
		panic(err)
	}

	for u := range ch {
		if u == nil {
			fmt.Println("snapshot done")
			continue
		}
		fmt.Println(u.Available, u.Name, u.ProviderID)
	}
}
