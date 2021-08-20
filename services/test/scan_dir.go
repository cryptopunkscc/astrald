package test

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/services/scanner"
	"github.com/cryptopunkscc/astrald/services/util/connect"
	"log"
	"time"
)

const pathToDir = "....."

func scanDir(ctx context.Context, core api.Core) {
	time.Sleep(2 * time.Second)

	conn, err := connect.Local(ctx, core, scanner.Port)
	defer conn.Close()
	if err != nil {
		return
	}

	log.Println(port, "writing path for scan")
	_, err = conn.WriteStringWithSize16(pathToDir)
	if err != nil {
		return
	}
	log.Println(port, "sent path for scan")
	for {
		_, err = conn.ReadByte()
		if err != nil {
			break
		}
	}
	log.Println(port, "finish path for scan")
}
