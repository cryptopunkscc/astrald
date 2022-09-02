package warpdrive

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/cslq"
	"io"
	"strings"
)

const prompt = "warp> "

func Cli(d *Dispatcher) (err error) {
	if !d.Authorized {
		return nil
	}
	prompt := prompt
	scanner := bufio.NewScanner(d.Conn)
	_, err = d.Conn.Write([]byte(prompt))
	if err != nil {
		err = Error(err, "Cannot write prompt")
		return
	}
	c, err := NewClient(d.Api).ConnectLocal()
	if err != nil {
		err = Error(err, "Cannot connect local client")
		return
	}
	finish := make(chan struct{})
	defer close(finish)
	go func() {
		select {
		case <-d.Done():
		case <-finish:
		}
		_ = d.Conn.Close()
		_ = c.Conn.Close()
	}()
	for scanner.Scan() {
		text := scanner.Text()
		switch text {
		case "prompt-off":
			prompt = ""
			_ = d.Encode("[c]c", "\n")
			continue
		case "e", "exit":
			return
		case "", "h", "help":
			_ = cliHelp(d.Context, d.Conn, c, nil)
			_ = d.Encode("[c]c", prompt)
			continue
		}

		words := strings.Split(text, " ")
		if len(words) == 0 {
			_ = d.Encode("[c]c", prompt)
			continue
		}
		cmd, args := words[0], words[1:]
		d.Logger = NewLogger(d.LogPrefix, fmt.Sprintf("(%s)", cmd))
		fn, ok := commands[cmd]
		if ok {
			err = fn(d.Context, d.Conn, c, args)
			if err != nil {
				err = Error(err, "cli command error")
				return
			}
			//d.Println("OK")
		} else {
			d.Println("no such cli command", cmd)
		}
		_ = d.Encode("[c]c", prompt)
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
type cmdFunc func(context.Context, io.ReadWriteCloser, LocalClient, []string) error

// =========================== Commands ===============================

func cliHelp(ctx context.Context, writer io.ReadWriteCloser, _ LocalClient, _ []string) (err error) {
	for name := range commands {
		if _, err = fmt.Fprintln(writer, name); err != nil {
			return err
		}
	}
	return
}

func cliPeers(ctx context.Context, writer io.ReadWriteCloser, client LocalClient, _ []string) (err error) {
	peers, err := client.Peers()
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

func cliSend(ctx context.Context, writer io.ReadWriteCloser, client LocalClient, args []string) (err error) {
	if len(args) < 1 {
		_, err = fmt.Fprintln(writer, "<filePath> <peerId>?")
		return
	}
	peer := client.LocalNode
	if len(args) > 1 {
		peer = args[1]
	}
	id, accepted, err := client.Send(PeerId(peer), args[0])
	if err != nil {
		return err
	}
	status := "delivered"
	if accepted {
		status = "accepted"
	}
	_, err = fmt.Fprintln(writer, id, status)
	return
}

func cliSent(ctx context.Context, writer io.ReadWriteCloser, client LocalClient, _ []string) (err error) {
	sent, err := client.Offers(FilterOut)
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

func cliReceived(ctx context.Context, writer io.ReadWriteCloser, client LocalClient, _ []string) (err error) {
	received, err := client.Offers(FilterIn)
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

func cliSubscribe(ctx context.Context, conn io.ReadWriteCloser, client LocalClient, args []string) (err error) {
	filter := "all"
	if len(args) > 0 {
		filter = args[0]
	}
	offers := make(<-chan Offer)
	switch filter {
	case "all", "out", "in":
		offers, err = client.Subscribe(Filter(filter))
	default:
		_, err = fmt.Fprintln(conn, "Invalid filter: ", filter)
		return
	}
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

func cliStatus(ctx context.Context, conn io.ReadWriteCloser, client LocalClient, args []string) (err error) {
	filter := "all"
	if len(args) > 0 {
		filter = args[0]
	}
	var events <-chan OfferStatus
	switch filter {
	case "all", "out", "in":
		events, err = client.Status(Filter(filter))
	default:
		_, err = fmt.Fprintln(conn, "Invalid filter: ", filter)
		return
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
	for event := range events {
		_, _ = fmt.Fprintln(conn, event.Id, event.Update, event.In, event.Status, event.Index, event.Progress)
	}
	return
}

func cliDownload(ctx context.Context, writer io.ReadWriteCloser, client LocalClient, args []string) (err error) {
	if len(args) < 1 {
		_, err = fmt.Fprintln(writer, "<offerId>")
		return
	}
	err = client.Accept(OfferId(args[0]))
	if err != nil {
		return
	}
	_, err = fmt.Fprintln(writer, "accepted")
	return
}

func cliUpdate(ctx context.Context, writer io.ReadWriteCloser, client LocalClient, args []string) (err error) {
	if len(args) < 3 {
		_, err = fmt.Fprintln(writer, "<peerId> <attr> <value>")
		return
	}
	err = client.Update(PeerId(args[0]), args[1], args[2])
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
