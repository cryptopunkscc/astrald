package warpdrive

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"io"
	"strings"
)

const prompt = "warp> "

func Cli(d *Dispatcher) (err error) {
	if !d.authorized {
		return nil
	}
	prompt := prompt
	scanner := bufio.NewScanner(d.conn)
	_, err = d.conn.Write([]byte(prompt))
	if err != nil {
		err = Error(err, "Cannot write prompt")
		return
	}
	c, err := NewClient(d.api).Connect(id.Identity{}, Port)
	if err != nil {
		err = Error(err, "Cannot connect local client")
		return
	}
	finish := make(chan struct{})
	defer close(finish)
	go func() {
		select {
		case <-d.ctx.Done():
		case <-finish:
		}
		_ = d.conn.Close()
		_ = c.conn.Close()
	}()
	for scanner.Scan() {
		text := scanner.Text()
		switch text {
		case "prompt-off":
			prompt = ""
			_ = d.cslq.Encode("[c]c", "\n")
			continue
		case "e", "exit":
			return
		case "", "h", "help":
			_ = cliHelp(d.ctx, d.conn, c, nil)
			_ = d.cslq.Encode("[c]c", prompt)
			continue
		}

		words := strings.Split(text, " ")
		if len(words) == 0 {
			_ = d.cslq.Encode("[c]c", prompt)
			continue
		}
		cmd, args := words[0], words[1:]
		d.log = NewLogger(d.logPrefix, fmt.Sprintf("(%s)", cmd))
		fn, ok := commands[cmd]
		if ok {
			err = fn(d.ctx, d.conn, c, args)
			if err != nil {
				err = Error(err, "cli command error")
				return
			}
			//d.Println("OK")
		} else {
			d.log.Println("no such cli command", cmd)
		}
		_ = d.cslq.Encode("[c]c", prompt)
	}
	return scanner.Err()
}

var commands = cmdMap{
	"peers":  cliPeers,
	"send":   cliSend,
	"out":    cliSent,
	"in":     cliReceived,
	"sub":    cliSubscribe,
	"get":    cliDownload,
	"update": cliUpdate,
	"stat":   cliStatus,
}

type cmdMap map[string]cmdFunc
type cmdFunc func(context.Context, io.ReadWriteCloser, Client, []string) error

var filters = map[string]Filter{
	"all": FilterAll,
	"in":  FilterIn,
	"out": FilterOut,
}

// =========================== Commands ===============================

func cliHelp(ctx context.Context, writer io.ReadWriteCloser, _ Client, _ []string) (err error) {
	for name := range commands {
		if _, err = fmt.Fprintln(writer, name); err != nil {
			return err
		}
	}
	return
}

func cliPeers(ctx context.Context, writer io.ReadWriteCloser, client Client, _ []string) (err error) {
	peers, err := client.ListPeers()
	if err != nil {
		return
	}
	for _, peer := range peers {
		_, err = fmt.Fprintln(writer, peer.Id, peer.Alias, peer.Mod)
		if err != nil {
			return
		}
	}
	return
}

func cliSend(ctx context.Context, writer io.ReadWriteCloser, client Client, args []string) (err error) {
	if len(args) < 1 {
		_, err = fmt.Fprintln(writer, "<filePath> <peerId>?")
		return
	}
	peer := client.localNode
	if len(args) > 1 {
		peer = args[1]
	}
	peerId, accepted, err := client.CreateOffer(PeerId(peer), args[0])
	if err != nil {
		return err
	}
	status := "delivered"
	if accepted {
		status = "accepted"
	}
	_, err = fmt.Fprintln(writer, peerId, status)
	return
}

func cliSent(ctx context.Context, writer io.ReadWriteCloser, client Client, _ []string) (err error) {
	sent, err := client.ListOffers(FilterOut)
	if err != nil {
		return err
	}
	for _, offer := range sent {
		err = printFilesRequest(writer, offer)
		if err != nil {
			return
		}
	}
	return
}

func cliReceived(ctx context.Context, writer io.ReadWriteCloser, client Client, _ []string) (err error) {
	received, err := client.ListOffers(FilterIn)
	if err != nil {
		return err
	}
	for _, offer := range received {
		err = printFilesRequest(writer, offer)
		if err != nil {
			return
		}
	}
	return
}

func cliSubscribe(ctx context.Context, conn io.ReadWriteCloser, client Client, args []string) (err error) {
	filter := "all"
	if len(args) > 0 {
		filter = args[0]
	}
	f, exist := filters[filter]
	if !exist {
		_, err = fmt.Fprintln(conn, "Invalid filter: ", filter)
		return
	}
	var offers <-chan Offer
	offers, err = client.ListenOffers(f)
	if err != nil {
		return err
	}
	finish := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
		case <-finish:
		}
		_ = client.Close()
		_ = conn.Close()
	}()
	go func() {
		defer close(finish)
		var code byte
		err = cslq.Decode(conn, "c", &code)
		if errors.Is(err, io.EOF) {
			err = nil
		}
	}()
	for offer := range offers {
		_ = printFilesRequest(conn, offer)
	}
	return
}

func cliStatus(ctx context.Context, conn io.ReadWriteCloser, client Client, args []string) (err error) {
	filter := "all"
	if len(args) > 0 {
		filter = args[0]
	}
	f, exist := filters[filter]
	if !exist {
		_, err = fmt.Fprintln(conn, "Invalid filter: ", filter)
		return
	}
	var events <-chan OfferStatus
	events, err = client.ListenStatus(f)
	finish := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
		case <-finish:
		}
		_ = client.Close()
		_ = conn.Close()
	}()
	go func() {
		defer close(finish)
		var code byte
		err = cslq.Decode(conn, "c", &code)
		if errors.Is(err, io.EOF) {
			err = nil
		}
	}()
	for event := range events {
		_, _ = fmt.Fprintln(conn, event.Id, event.Update, event.In, event.Status, event.Index, event.Progress)
	}
	return
}

func cliDownload(ctx context.Context, writer io.ReadWriteCloser, client Client, args []string) (err error) {
	if len(args) < 1 {
		_, err = fmt.Fprintln(writer, "<offerId>")
		return
	}
	err = client.AcceptOffer(OfferId(args[0]))
	if err != nil {
		return
	}
	_, err = fmt.Fprintln(writer, "accepted")
	return
}

func cliUpdate(ctx context.Context, writer io.ReadWriteCloser, client Client, args []string) (err error) {
	if len(args) < 3 {
		_, err = fmt.Fprintln(writer, "<peerId> <attr> <value>")
		return
	}
	err = client.UpdatePeer(PeerId(args[0]), args[1], args[2])
	if err != nil {
		return
	}
	_, err = fmt.Fprintln(writer, "updated")
	return
}

func printFilesRequest(writer io.Writer, offer Offer) (err error) {
	_, err = fmt.Fprintln(writer, "incoming:", offer.In)
	_, err = fmt.Fprintln(writer, "peer:", offer.Peer)
	_, err = fmt.Fprintln(writer, "offer id:", offer.Id)
	_, err = fmt.Fprintln(writer, "created at:", offer.Create)
	_, err = fmt.Fprintln(writer, "status:", offer.Status)
	if offer.Index > -1 {
		_, err = fmt.Fprintln(writer, "  file index:", offer.Index)
		_, err = fmt.Fprintln(writer, "  progress:", offer.Progress)
		_, err = fmt.Fprintln(writer, "  update at:", offer.Update)
	}
	_, err = fmt.Fprintln(writer, "files:")
	if err != nil {
		return
	}
	for i, file := range offer.Files {
		_, err = fmt.Fprintf(writer, "%d. %s (%dB)\n", i, file.Name, file.Size)
		if err != nil {
			return
		}
	}
	_, err = fmt.Fprintln(writer, "-----------------------------------")
	return
}
