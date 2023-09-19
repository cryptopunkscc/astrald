package admin

import (
	"errors"
	"flag"
	"github.com/cryptopunkscc/astrald/node/services"
	"sort"
	"strings"
	"time"
)

const timestampFormat string = "2006-01-02 15:04:05"

type CmdServices struct {
	mod *Module
}

func (cmd *CmdServices) Exec(term *Terminal, args []string) error {
	if len(args) < 2 {
		return cmd.help(term)
	}

	switch args[1] {
	case "list":
		return cmd.list(term, args[2:])

	case "help":
		return cmd.help(term)

	case "release":
		return cmd.release(term, args[2:])

	default:
		return errors.New("invalid command")
	}
}

func (cmd *CmdServices) list(term *Terminal, args []string) error {
	var sortBy string
	var showTime bool

	var f = flag.NewFlagSet("list", flag.ContinueOnError)
	f.SetOutput(term)
	f.StringVar(&sortBy, "s", "name", "sort by name/age/identity")
	f.BoolVar(&showTime, "t", false, "show registration time instead of age")
	if err := f.Parse(args); err != nil {
		return nil
	}

	var list = cmd.mod.node.Services().List()

	sorter := serviceSorter{list: list}
	switch sortBy[0] {
	case 'i':
		sorter.Criteria = byIdentity
	case 'a':
		sorter.Criteria = byAge
	default:
		sorter.Criteria = byName
	}

	sort.Sort(sorter)

	var format = "%-40s %-12s %s\n"

	if showTime {
		format = "%-40s %-21s %s\n"
		term.Printf(format, Header("NAME"), Header("TIME"), Header("IDENTITY"))
	} else {
		term.Printf(format, Header("NAME"), Header("AGE"), Header("IDENTITY"))
	}

	for _, service := range list {
		var age any = time.Since(service.RegisteredAt).Round(time.Second)

		if showTime {
			age = service.RegisteredAt
		}

		term.Printf(format, Keyword(service.Name), age, service.Identity)
	}
	return nil
}

func (cmd *CmdServices) release(term *Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("missing argument")
	}

	for _, service := range cmd.mod.node.Services().FindByName(args[0]) {
		service.Close()
	}

	return nil
}

func (cmd *CmdServices) help(term *Terminal) error {
	term.Printf("help: tracker <command> [options]\n\n")
	term.Printf("commands:\n")
	term.Printf("  list              list all services\n")
	term.Printf("  release <name>    release a service\n")
	term.Printf("  help              show help\n")
	return nil
}

func (cmd *CmdServices) ShortDescription() string {
	return "view and manage local services"
}

type serviceSorter struct {
	list     []services.ServiceInfo
	Criteria func(services.ServiceInfo, services.ServiceInfo) bool
}

func (s serviceSorter) Len() int {
	return len(s.list)
}

func (s serviceSorter) Less(i, j int) bool {
	if s.Criteria != nil {
		return s.Criteria(s.list[i], s.list[j])
	}
	return strings.Compare(s.list[i].Name, s.list[j].Name) == -1
}

func (s serviceSorter) Swap(i, j int) {
	s.list[i], s.list[j] = s.list[j], s.list[i]
}

func byAge(a services.ServiceInfo, b services.ServiceInfo) bool {
	return a.RegisteredAt.Before(b.RegisteredAt)
}

func byName(a services.ServiceInfo, b services.ServiceInfo) bool {
	return strings.Compare(a.Name, b.Name) == -1
}

func byIdentity(a services.ServiceInfo, b services.ServiceInfo) bool {
	return strings.Compare(a.Identity.String(), b.Identity.String()) == -1
}
