package content

import (
	"encoding/json"
	"errors"
	"flag"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/object"
	"time"
)

type Admin struct {
	mod  *Module
	cmds map[string]func(admin.Terminal, []string) error
}

func NewAdmin(mod *Module) *Admin {
	var cmd = &Admin{mod: mod}
	cmd.cmds = map[string]func(admin.Terminal, []string) error{
		"scan":      cmd.scan,
		"identify":  cmd.identify,
		"forget":    cmd.forget,
		"set_label": cmd.setLabel,
		"get_label": cmd.getLabel,
	}
	return cmd
}

func (cmd *Admin) scan(term admin.Terminal, args []string) error {
	var opts = &content.ScanOpts{}
	var since string

	var flags = flag.NewFlagSet("scan", flag.ContinueOnError)
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

	list, err := cmd.mod.scan(opts)
	if err != nil {
		return err
	}

	var format = "%-64s %-8s %-32s %s\n"
	term.Printf(format, admin.Header("ID"), admin.Header("Method"), admin.Header("Type"), admin.Header("Label"))
	for _, item := range list {
		term.Printf(format,
			item.ObjectID,
			item.Method,
			item.Type,
			cmd.mod.GetLabel(item.ObjectID),
		)
	}

	return nil
}

func (cmd *Admin) identify(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("missing argument")
	}

	objectID, err := object.ParseID(args[0])
	if err != nil {
		return err
	}

	info, err := cmd.mod.Identify(objectID)
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

func (cmd *Admin) forget(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("missing argument")
	}

	objectID, err := object.ParseID(args[0])
	if err != nil {
		return err
	}

	return cmd.mod.Forget(objectID)
}

func (cmd *Admin) setLabel(term admin.Terminal, args []string) error {
	if len(args) < 2 {
		return errors.New("missing argument")
	}

	objectID, err := object.ParseID(args[0])
	if err != nil {
		return err
	}

	label := args[1]

	cmd.mod.SetLabel(objectID, label)

	return nil
}

func (cmd *Admin) getLabel(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("missing argument")
	}

	objectID, err := object.ParseID(args[0])
	if err != nil {
		return err
	}

	term.Printf("%s\n", cmd.mod.GetLabel(objectID))
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
	term.Printf("  scan [args]                    list identified objects\n")
	term.Printf("  identify <objectID>            identify an object\n")
	term.Printf("  forget <objectID>              forget an object (remove from cache)\n")
	term.Printf("  set_label <objectID> <label>   assign a label to an object\n")
	term.Printf("  get_label <objectID>           show object's label\n")
	term.Printf("  help                           show help\n")
	return nil
}

func (cmd *Admin) ShortDescription() string {
	return "content identification"
}
