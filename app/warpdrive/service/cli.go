package warpdrive

import (
	"bufio"
	"fmt"
	astral "github.com/cryptopunkscc/astrald/mod/apphost/client"
	"io"
	"log"
	"strings"
)

const prompt = "warp> "
const cliPort = "wd"

func (srv *service) handleCommandLine() {
	port := srv.register(cliPort)
	for request := range port.Next() {
		go func(request *astral.Request) {
			if srv.isRejected(request) {
				return
			}
			conn, err := request.Accept()
			if err != nil {
				log.Println(">", ACCEPT, "Cannot accept", err)
				return
			}
			go serve(conn)
		}(request)
	}
}

func serve(stream io.ReadWriteCloser) {
	defer stream.Close()
	scanner := bufio.NewScanner(stream)
	stream.Write([]byte(prompt))
	api := NewUIClient()
	for scanner.Scan() {
		words := strings.Split(scanner.Text(), " ")
		if len(words) == 0 {
			continue
		}
		cmd, args := words[0], words[1:]
		fn, ok := commands[cmd]
		if ok {
			err := fn(stream, api, args)
			if err != nil {
				fmt.Fprintf(stream, "error: %v\n", err)
			} else {
				fmt.Fprintf(stream, "ok\n")
			}
		} else {
			fmt.Fprintf(stream, "no such command\n")
		}
		stream.Write([]byte(prompt))
	}
}

func init() {
	commands = cmdMap{
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
}

var commands cmdMap

type cmdMap map[string]cmdFunc
type cmdFunc func(io.ReadWriter, ClientApi, []string) error

// =========================== Commands ===============================

func cmdPeers(writer io.ReadWriter, api ClientApi, _ []string) (err error) {
	peers, err := api.Peers()
	if err != nil {
		return
	}
	log.Println("peers:", peers)
	for _, peer := range peers {
		_, err = fmt.Fprintln(writer, peer.Id, peer.Alias, peer.Mod)
		if err != nil {
			return
		}
	}
	return
}

func cmdSend(writer io.ReadWriter, api ClientApi, args []string) (err error) {
	peer := ""
	if len(args) > 1 {
		peer = args[1]
	}
	id, err := api.Send(PeerId(peer), args[0])
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(writer, id)
	return
}

func cmdStatus(writer io.ReadWriter, api ClientApi, args []string) (err error) {
	status, err := api.Status(OfferId(args[0]))
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(writer, status)
	return
}

func cmdSent(writer io.ReadWriter, api ClientApi, _ []string) (err error) {
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

func cmdReceived(writer io.ReadWriter, api ClientApi, _ []string) (err error) {
	received, err := api.Received("")
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

func printFilesRequest(writer io.Writer, offer Offer) (err error) {
	_, err = fmt.Fprintln(writer, offer.Id, offer.Peer, offer.Status.Status)
	if err != nil {
		return
	}
	for _, file := range offer.Files {
		_, err = fmt.Fprintln(writer, "  - ", file.Path)
		if err != nil {
			return
		}
	}
	return
}

func cmdOffers(writer io.ReadWriter, api ClientApi, _ []string) (err error) {
	offers, err := api.Offers()
	if err != nil {
		return err
	}
	for offer := range offers {
		_ = printFilesRequest(writer, offer)
	}
	return
}

func cmdAccept(writer io.ReadWriter, api ClientApi, args []string) (err error) {
	err = api.Accept(OfferId(args[0]))
	if err != nil {
		return
	}
	_, err = fmt.Fprintln(writer, "accepted", args[0])
	return
}

func cmdReject(writer io.ReadWriter, api ClientApi, args []string) (err error) {
	err = api.Reject(OfferId(args[0]))
	if err != nil {
		return
	}
	_, err = fmt.Fprintln(writer, "rejected", args[0])
	return
}

func cmdUpdate(writer io.ReadWriter, api ClientApi, args []string) (err error) {
	err = api.Update(PeerId(args[0]), args[1], args[2])
	if err != nil {
		return
	}
	_, err = fmt.Fprintln(writer, "updated", args)
	return
}

func cmdEvents(writer io.ReadWriter, api ClientApi, args []string) (err error) {
	filter := "all"
	if len(args) > 0 {
		filter = args[0]
	}
	var events <-chan Status
	switch filter {
	case "all":
		events, err = api.Events()
	case "sent":
		events, err = api.Sender().Events()
	case "received":
		events, err = api.Recipient().Events()
	}
	for event := range events {
		_, _ = fmt.Fprintln(writer, event.Id, event.Status)
	}
	return
}
