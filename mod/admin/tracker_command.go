package admin

import (
	"errors"
	"flag"
	"fmt"
	"github.com/cryptopunkscc/astrald/node"
	"time"
)

var _ Command = &TrackerCommand{}

type TrackerCommand struct {
	flags *flag.FlagSet
	node  node.Node
}

func (cmd *TrackerCommand) Exec(t *Terminal, args []string) error {
	if len(args) < 2 {
		cmd.usage(t)
		return errors.New("missing command")
	}

	switch args[1] {
	case "list":
		return cmd.list(t)

	case "add":
		return cmd.add(t, args[2:])

	default:
		cmd.usage(t)
		return errors.New("unknown command")
	}
}

func (cmd *TrackerCommand) add(t *Terminal, args []string) error {
	if len(args) < 3 {
		t.Println("usage: tracker add <identity> <network> <address>")
		return errors.New("misisng arguments")
	}

	identity, err := cmd.node.Contacts().ResolveIdentity(args[0])

	network, address := args[1], args[2]

	endpoint, err := cmd.node.Infra().Parse(network, address)
	if err != nil {
		return err
	}

	duration := 30 * 24 * time.Hour

	if len(args) > 3 {
		duration, err = time.ParseDuration(args[3])
		if err != nil {
			return err
		}
	}

	return cmd.node.Tracker().Add(identity, endpoint, time.Now().Add(duration))
}

func (cmd *TrackerCommand) list(t *Terminal) error {
	ids, err := cmd.node.Tracker().Identities()
	if err != nil {
		return err
	}

	for _, nodeID := range ids {
		name := cmd.node.Contacts().DisplayName(nodeID)
		keyHex := nodeID.PublicKeyHex()

		t.Printf("%s (%s)\n", name, keyHex)

		addrs, err := cmd.node.Tracker().AddrByIdentity(nodeID)
		if err != nil {
			return err
		}

		for _, addr := range addrs {
			printEndpoint(t, addr)
		}

		fmt.Fprintln(t)
	}

	return nil
}

func (cmd *TrackerCommand) usage(t *Terminal) {
	t.Printf("usage: tracker <command> [options]\n\n")
	t.Printf("commands: list, add\n")
}

func (cmd *TrackerCommand) HelpDescription() string {
	return "manage node tracker entries"
}
