package test

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/services/repo"
	"github.com/cryptopunkscc/astrald/services/repo/request"
	"github.com/cryptopunkscc/astrald/services/util/connect"
	"log"
	"time"
)

const pathToFile = "....."

func mapFile(ctx context.Context, core api.Core) {
	time.Sleep(2 * time.Second)
	conn, err := connect.LocalRequest(ctx, core, repo.Port, request.Map)
	defer conn.Close()
	if err != nil {
		return
	}
	log.Println(port, "writing path to map")
	_, err = conn.WriteStringWithSize16(pathToFile)
	if err != nil {
		return
	}
	log.Println(port, "path sent to map")
	for {
		_, err = conn.ReadByte()
		if err != nil {
			break
		}
	}
	log.Println(port, "mapped successfully")
}
