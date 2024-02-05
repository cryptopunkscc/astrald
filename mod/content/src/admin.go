package content

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/sig"
	"slices"
	"time"
)

type Admin struct {
	mod  *Module
	cmds map[string]func(admin.Terminal, []string) error
}

func NewAdmin(mod *Module) *Admin {
	var cmd = &Admin{mod: mod}
	cmd.cmds = map[string]func(admin.Terminal, []string) error{
		"find":       cmd.find,
		"identify":   cmd.identify,
		"forget":     cmd.forget,
		"describe":   cmd.describe,
		"prototypes": cmd.prototypes,
		"set_label":  cmd.setLabel,
		"get_label":  cmd.getLabel,
	}
	return cmd
}

func (cmd *Admin) find(term admin.Terminal, args []string) error {
	var opts = &content.ScanOpts{}
	var since string

	var flags = flag.NewFlagSet("find", flag.ContinueOnError)
	flags.StringVar(&opts.Type, "t", "", "show objects of this type only")
	flags.StringVar(&since, "a", "", "show objects indexed after a time (YYYY-MM-DD HH:MM:SS)")
	flags.SetOutput(term)
	err := flags.Parse(args)
	if err != nil {
		return nil
	}

	if since != "" {
		opts.After, err = time.Parse(time.DateTime, since)
		if err != nil {
			return err
		}
	}

	list, err := cmd.mod.find(opts)
	if err != nil {
		return err
	}

	var format = "%-64s %-8s %-32s %s\n"
	term.Printf(format, admin.Header("ID"), admin.Header("Method"), admin.Header("Type"), admin.Header("Label"))
	for _, item := range list {
		term.Printf(format,
			item.DataID,
			item.Method,
			item.Type,
			cmd.mod.GetLabel(item.DataID),
		)
	}

	return nil
}

func (cmd *Admin) identify(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("missing argument")
	}

	dataID, err := data.Parse(args[0])
	if err != nil {
		info, err := cmd.mod.IdentifySet(args[0])
		if err != nil {
			return err
		}

		j, err := json.MarshalIndent(info, "", "  ")
		if err != nil {
			return err
		}

		term.Write(j)
		term.Println()

		return nil
	}

	info, err := cmd.mod.Identify(dataID)
	if err != nil {
		return err
	}

	j, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return err
	}

	term.Write(j)
	term.Println()

	return nil
}

func (cmd *Admin) describe(term admin.Terminal, args []string) error {
	var err error
	var opts = &content.DescribeOpts{
		IdentityFilter: id.AllowEveryone,
	}

	var flags = flag.NewFlagSet("describe", flag.ContinueOnError)
	flags.BoolVar(&opts.Network, "n", false, "use network sources")
	flags.SetOutput(term)
	err = flags.Parse(args)
	if err != nil {
		return err
	}

	if len(flags.Args()) == 0 {
		return errors.New("missing data id")
	}

	dataID, err := data.Parse(args[0])
	if err != nil {
		return err
	}

	var desc = cmd.mod.Describe(context.Background(), dataID, opts)

	for _, d := range desc {
		term.Printf("%v: %v\n  ", d.Source, admin.Keyword(d.Data.DescriptorType()))

		j, err := json.MarshalIndent(d.Data, "  ", "  ")
		if err != nil {
			term.Printf("marshal error: %v\n", err)
		}
		term.Printf("%s\n\n", string(j))
	}

	return nil
}

func (cmd *Admin) forget(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("missing argument")
	}

	dataID, err := data.Parse(args[0])
	if err != nil {
		return err
	}

	return cmd.mod.Forget(dataID)
}

func (cmd *Admin) prototypes(term admin.Terminal, args []string) error {
	list, _ := sig.MapSlice(cmd.mod.prototypes.Values(), func(i content.DescriptorData) (string, error) {
		return i.DescriptorType(), nil
	})

	slices.Sort(list)

	for _, p := range list {
		term.Printf("%s\n", p)
	}
	return nil
}

func (cmd *Admin) setLabel(term admin.Terminal, args []string) error {
	if len(args) < 2 {
		return errors.New("missing argument")
	}

	dataID, err := data.Parse(args[0])
	if err != nil {
		return err
	}

	label := args[1]

	cmd.mod.SetLabel(dataID, label)

	return nil
}

func (cmd *Admin) getLabel(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("missing argument")
	}

	dataID, err := data.Parse(args[0])
	if err != nil {
		return err
	}

	term.Printf("%s\n", cmd.mod.GetLabel(dataID))
	return nil
}

func (cmd *Admin) Exec(term admin.Terminal, args []string) error {
	if len(args) < 2 {
		return cmd.help(term, []string{})
	}

	c, args := args[1], args[2:]
	if fn, found := cmd.cmds[c]; found {
		return fn(term, args)
	}

	return errors.New("unknown command")
}

func (cmd *Admin) help(term admin.Terminal, _ []string) error {
	term.Printf("usage: %s <command>\n\n", content.ModuleName)
	term.Printf("commands:\n")
	term.Printf("  find [args]                  list identified objects\n")
	term.Printf("  identify <dataID|set>        identify an object's type\n")
	term.Printf("  forget <dataID>              forget an object (remove from cache)\n")
	term.Printf("  describe <dataID>            describe an object\n")
	term.Printf("  set_label <dataID> <label>   assign a label to an object\n")
	term.Printf("  get_label <dataID>           show object's label\n")
	term.Printf("  help                         show help\n")
	return nil
}

func (cmd *Admin) ShortDescription() string {
	return "content identification"
}
