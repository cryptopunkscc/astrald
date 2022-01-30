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

func CommandLine(srv handler.Context, request astral.Request) {
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
	"peers":    cmdPeers,
	"send":     cmdSend,
	"status":   cmdStatus,
	"sent":     cmdSent,
	"received": cmdReceived,
	"offers":   cmdOffers,
	"accept":   cmdAccept,
	"reject":   cmdReject,
	"update":   cmdUpdate,
	"events":   cmdEvents,
}

type cmdMap map[string]cmdFunc
type cmdFunc func(io.ReadWriter, Client, []string) error

// =========================== Commands ===============================

func cmdPeers(writer io.ReadWriter, api Client, _ []string) (err error) {
	peers, err := api.Peers()
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
		_, err = fmt.Fprintln(writer, "<peerId> <filePath>?")
		return
	}
	peer := ""
	if len(args) > 1 {
		peer = args[1]
	}
	id, err := client.Send(api.PeerId(peer), args[0])
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(writer, id)
	return
}

func cmdStatus(writer io.ReadWriter, client Client, args []string) (err error) {
	if len(args) < 1 {
		_, err = fmt.Fprintln(writer, "<offerId>")
		return
	}
	status, err := client.Status(api.OfferId(args[0]))
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(writer, status)
	return
}

func cmdSent(writer io.ReadWriter, api Client, _ []string) (err error) {
	sent, err := api.Sent()
	if err != nil {
		return err
	}
	for _, offer := range sent {
		err = printFilesRequest(writer, *offer)
		if err != nil {
			return
		}
	}
	return
}

func cmdReceived(writer io.ReadWriter, api Client, _ []string) (err error) {
	received, err := api.Received()
	if err != nil {
		return err
	}
	for _, offer := range received {
		err = printFilesRequest(writer, *offer)
		if err != nil {
			return
		}
	}
	return
}

func printFilesRequest(writer io.Writer, offer api.Offer) (err error) {
	_, err = fmt.Fprintln(writer, offer.Id, offer.Peer, offer.Status.Status)
	if err != nil {
		return
	}
	for _, file := range offer.Files {
		_, err = fmt.Fprintln(writer, "  - ", file.Uri)
		if err != nil {
			return
		}
	}
	return
}

func cmdOffers(writer io.ReadWriter, api Client, _ []string) (err error) {
	offers, err := api.Offers()
	if err != nil {
		return err
	}
	for offer := range offers {
		_ = printFilesRequest(writer, offer)
	}
	return
}

func cmdAccept(writer io.ReadWriter, client Client, args []string) (err error) {
	if len(args) < 1 {
		_, err = fmt.Fprintln(writer, "<offerId>")
		return
	}
	err = client.Accept(api.OfferId(args[0]))
	if err != nil {
		return
	}
	_, err = fmt.Fprintln(writer, "accepted")
	return
}

func cmdReject(writer io.ReadWriter, client Client, args []string) (err error) {
	if len(args) < 1 {
		_, err = fmt.Fprintln(writer, "<offerId>")
		return
	}
	err = client.Reject(api.OfferId(args[0]))
	if err != nil {
		return
	}
	_, err = fmt.Fprintln(writer, "rejected")
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

func cmdEvents(writer io.ReadWriter, client Client, args []string) (err error) {
	filter := "all"
	if len(args) > 0 {
		filter = args[0]
	}
	var events <-chan api.Status
	switch filter {
	case "all":
		events, err = client.Events()
	case "sent":
		events, err = client.Sender.Events()
	case "received":
		events, err = client.Recipient.Events()
	}
	for event := range events {
		_, _ = fmt.Fprintln(writer, event.Id, event.Status)
	}
	return
}
