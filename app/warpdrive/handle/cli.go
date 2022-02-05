package handle

import (
	"bufio"
	"fmt"
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"github.com/cryptopunkscc/astrald/app/warpdrive/handler"
	astral "github.com/cryptopunkscc/astrald/mod/apphost/client"
	"io"
	"strings"
)

const prompt = "warp> "

func Cli(srv handler.Context, request astral.Request) {
	if srv.IsRejected(request) {
		return
	}
	conn, err := request.Accept()
	if err != nil {
		srv.Println("Cannot accept", err)
		return
	}
	go serve(srv, conn)
}

func serve(srv handler.Context, stream io.ReadWriteCloser) {
	defer stream.Close()
	scanner := bufio.NewScanner(stream)
	_, err := stream.Write([]byte(prompt))
	if err != nil {
		srv.Panicln("Cannot write prompt", err)
	}
	c := NewClient(srv.Api)
	for scanner.Scan() {
		words := strings.Split(scanner.Text(), " ")
		if len(words) == 0 {
			continue
		}
		cmd, args := words[0], words[1:]
		fn, ok := commands[cmd]
		if ok {
			err := fn(stream, c, args)
			if err != nil {
				srv.Println("command error", err)
			}
		} else {
			srv.Println("no such command", cmd)
		}
		_, _ = stream.Write([]byte(prompt))
	}
}

var commands = cmdMap{
	"peers":  cmdPeers,
	"send":   cmdSend,
	"out":    cmdSent,
	"in":     cmdReceived,
	"sub":    cmdSubscribe,
	"get":    cmdDownload,
	"update": cmdUpdate,
	"stat":   cmdStatus,
}

type cmdMap map[string]cmdFunc
type cmdFunc func(io.ReadWriter, Client, []string) error

// =========================== Commands ===============================

func cmdPeers(writer io.ReadWriter, client Client, _ []string) (err error) {
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

func cmdSend(writer io.ReadWriter, client Client, args []string) (err error) {
	if len(args) < 1 {
		_, err = fmt.Fprintln(writer, "<filePath> <peerId>?")
		return
	}
	peer := ""
	if len(args) > 1 {
		peer = args[1]
	}
	id, accepted, err := client.Send(api.PeerId(peer), args[0])
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

func cmdSent(writer io.ReadWriter, client Client, _ []string) (err error) {
	sent, err := client.Offers(api.FilterOut)
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

func cmdReceived(writer io.ReadWriter, client Client, _ []string) (err error) {
	received, err := client.Offers(api.FilterIn)
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

func cmdSubscribe(writer io.ReadWriter, client Client, args []string) (err error) {
	filter := "all"
	if len(args) > 0 {
		filter = args[0]
	}
	offers := make(<-chan api.Offer)
	switch filter {
	case "all", "out", "in":
		offers, err = client.Subscribe(api.Filter(filter))
	default:
		_, err = fmt.Fprintln(writer, "Invalid filter: ", filter)
		return
	}
	if err != nil {
		return err
	}
	for offer := range offers {
		_ = printFilesRequest(writer, offer)
	}
	return
}

func cmdDownload(writer io.ReadWriter, client Client, args []string) (err error) {
	if len(args) < 1 {
		_, err = fmt.Fprintln(writer, "<offerId>")
		return
	}
	err = client.Download(api.OfferId(args[0]))
	if err != nil {
		return
	}
	_, err = fmt.Fprintln(writer, "accepted")
	return
}

func cmdUpdate(writer io.ReadWriter, client Client, args []string) (err error) {
	if len(args) < 3 {
		_, err = fmt.Fprintln(writer, "<peerId> <attr> <value>")
		return
	}
	err = client.Update(api.PeerId(args[0]), args[1], args[2])
	if err != nil {
		return
	}
	_, err = fmt.Fprintln(writer, "updated")
	return
}

func cmdStatus(writer io.ReadWriter, client Client, args []string) (err error) {
	filter := "all"
	if len(args) > 0 {
		filter = args[0]
	}
	var events <-chan api.Status
	switch filter {
	case "all", "out", "in":
		events, err = client.Status(api.Filter(filter))
	default:
		_, err = fmt.Fprintln(writer, "Invalid filter: ", filter)
		return
	}
	for event := range events {
		_, _ = fmt.Fprintln(writer, event.Id, event.Status)
	}
	return
}

func printFilesRequest(writer io.Writer, offer api.Offer) (err error) {
	_, err = fmt.Fprintln(writer, "incoming:", offer.In)
	_, err = fmt.Fprintln(writer, "peer:", offer.Peer)
	_, err = fmt.Fprintln(writer, "offer id:", offer.Id)
	_, err = fmt.Fprintln(writer, "created at:", offer.Create)
	_, err = fmt.Fprintln(writer, "status:", offer.Status.Status)
	if offer.Index > -1 {
		_, err = fmt.Fprintln(writer, "  file index:", offer.Index)
		_, err = fmt.Fprintln(writer, "  progress:", offer.Progress)
		_, err = fmt.Fprintln(writer, "  update at:", offer.Progress)
	}
	_, err = fmt.Fprintln(writer, "files:")
	if err != nil {
		return
	}
	for i, file := range offer.Files {
		_, err = fmt.Fprintln(writer, i, "-", file.Uri, file.Size)
		if err != nil {
			return
		}
	}
	_, err = fmt.Fprintln(writer, "-----------------------------------")
	return
}
