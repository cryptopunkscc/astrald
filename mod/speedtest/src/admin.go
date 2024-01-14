package speedtest

import (
	"context"
	"fmt"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/streams"
	"io"
	"strconv"
	"time"
)

type Admin struct {
	mod *Module
}

func NewAdmin(mod *Module) *Admin {
	var adm = &Admin{mod: mod}

	return adm
}

func (adm *Admin) Exec(term admin.Terminal, args []string) error {
	if len(args) < 2 {
		return adm.help(term, []string{})
	}

	linkID, err := strconv.Atoi(args[1])
	if err != nil {
		return err
	}

	link, err := adm.mod.node.Network().Links().Find(linkID)
	if err != nil {
		return err
	}

	query := net.NewQuery(adm.mod.node.Identity(), link.RemoteIdentity(), ServiceName)
	conn, err := net.Route(context.Background(), link, query)
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := cslq.Encode(conn, "c", 10); err != nil {
		return err
	}

	var errCode int
	cslq.Decode(conn, "c", &errCode)
	if errCode != 0 {
		return fmt.Errorf("speedtest service returned error code %v", errCode)
	}

	term.Printf("running speedtest...\n")

	var read int64
	var startAt = time.Now()

	read, _ = io.Copy(streams.NilWriter{}, conn)

	var elapsed = float64(time.Since(startAt)) / float64(time.Second)
	var speed = int(float64(read) / elapsed)

	term.Printf("read %v in %v seconds, speed %v/s\n",
		log.DataSize(read).HumanReadable(),
		elapsed,
		log.DataSize(speed).HumanReadable(),
	)

	return nil
}

func (adm *Admin) help(term admin.Terminal, _ []string) error {
	term.Printf("usage: speedtest <linkID>\n")
	return nil
}

func (adm *Admin) ShortDescription() string {
	return "run a speedtest"
}
