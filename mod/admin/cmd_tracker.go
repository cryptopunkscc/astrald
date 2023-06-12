package admin

import (
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/infra/drivers/gw"
	"github.com/cryptopunkscc/astrald/node/tracker"
	"time"
)

const defaultAddDuration = 30 * 24 * time.Hour

var _ Command = &CmdTracker{}

type CmdTracker struct {
	mod *Module
}

func (cmd *CmdTracker) Exec(term *Terminal, args []string) error {
	if len(args) < 2 {
		return cmd.help(term)
	}

	switch args[1] {
	case "list":
		return cmd.list(term)

	case "add":
		return cmd.add(term, args[2:])

	case "set_alias":
		return cmd.setAlias(term, args[2:])

	case "show":
		return cmd.show(term, args[2:])

	case "help":
		return cmd.help(term)

	case "delete":
		return cmd.delete(term, args[2:])

	default:
		return errors.New("invalid command")
	}
}

func (cmd *CmdTracker) add(term *Terminal, args []string) error {
	if len(args) < 3 {
		term.Println("usage: tracker add <node> <network> <address>")
		return errors.New("misisng arguments")
	}

	identity, err := cmd.mod.node.Resolver().Resolve(args[0])
	if err != nil {
		return err
	}

	ep, err := cmd.mod.node.Infra().Parse(args[1], args[2])
	if err != nil {
		return err
	}

	var duration = defaultAddDuration

	if len(args) > 3 {
		duration, err = time.ParseDuration(args[3])
		if err != nil {
			return err
		}
	}

	return cmd.mod.node.Tracker().AddEndpoint(identity, ep, time.Now().Add(duration))
}

func (cmd *CmdTracker) setAlias(term *Terminal, args []string) error {
	if len(args) < 2 {
		return errors.New("not enough arguments")
	}

	identity, err := cmd.mod.node.Resolver().Resolve(args[0])
	if err != nil {
		return err
	}

	return cmd.mod.node.Tracker().SetAlias(identity, args[1])
}

func (cmd *CmdTracker) show(term *Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("not enough arguments")
	}

	identity, err := cmd.mod.node.Resolver().Resolve(args[0])
	if err != nil {
		return err
	}

	term.Printf("%s (%s)\n", identity, Faded(identity.String()))

	// check private key
	if keys, err := cmd.mod.assets.KeyStore(); err == nil {
		if pi, err := keys.Find(identity); err == nil {
			if pi.PrivateKey() != nil {
				term.Printf("%s\n", Important("private key available"))
			}
		}
	}

	term.Println()

	// print endpoints
	if endpoints, err := cmd.mod.node.Tracker().EndpointsByIdentity(identity); err == nil {
		if len(endpoints) == 0 {
			term.Printf("no known endpoints.\n")
		} else {
			var f = "%-10s %-40s %s\n"
			term.Printf(f, Header("Network"), Header("Address"), Header("Expires"))
			for _, ep := range endpoints {
				term.Printf(f, ep.Network(), ep.Endpoint, ep.ExpiresAt)
			}
			term.Printf("%d %s\n", len(endpoints), Faded("endpoints."))
		}
	}

	return nil
}

func (cmd *CmdTracker) list(term *Terminal) error {
	ids, err := cmd.mod.node.Tracker().Identities()
	if err != nil {
		return err
	}

	var f = "%-30s %s\n"
	term.Printf(f, Header("Alias"), Header("PubKey"))
	for _, nodeID := range ids {
		term.Printf(f, nodeID, nodeID.String())
	}

	return nil
}

func (cmd *CmdTracker) help(term *Terminal) error {
	term.Printf("help: tracker <command> [options]\n\n")
	term.Printf("commands:\n")
	term.Printf("  list                              list all identities\n")
	term.Printf("  add                               add an endpoint to an identity\n")
	term.Printf("  set_alias <identity> <alias>      set identity's alias\n")
	term.Printf("  show <identity>                   show info about an identity\n")
	term.Printf("  delete                            delete identity endpoints\n")
	term.Printf("  help                              show help\n")
	return nil
}

func (cmd *CmdTracker) delete(term *Terminal, args []string) error {
	if len(args) < 1 {
		term.Println("usage: tracker delete <node>")
		return errors.New("misisng arguments")
	}

	identity, err := cmd.mod.node.Resolver().Resolve(args[0])
	if err != nil {
		return err
	}

	return cmd.mod.node.Tracker().DeleteAll(identity)
}

func (cmd *CmdTracker) ShortDescription() string {
	return "manage node tracker entries"
}

func (cmd *CmdTracker) formatEndpoint(endpoint net.Endpoint) string {
	var suffix string

	if e, ok := endpoint.(tracker.TrackedEndpoint); ok {
		d := e.ExpiresAt.Sub(time.Now()).Round(time.Second)
		suffix = fmt.Sprintf(" (expires in %s)", d)
		endpoint = e.Endpoint
	}

	network, address := endpoint.Network(), endpoint.String()

	if e, ok := endpoint.(gw.Endpoint); ok {
		var r = cmd.mod.node.Resolver()
		address = fmt.Sprintf("%s:%s", r.DisplayName(e.Gate()), r.DisplayName(e.Target()))
	}

	return fmt.Sprintf("%-10s%s"+suffix, network, address)
}
