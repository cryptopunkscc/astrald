package setup

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
)

type SetupDialogue struct {
	*Module
	*Dialogue
}

func NewSetupDialogue(mod *Module, rw io.ReadWriter) *SetupDialogue {
	return &SetupDialogue{
		Module:   mod,
		Dialogue: NewDialogue(rw),
	}
}

func (d *SetupDialogue) start() error {
	d.Say("ASTRAL NODE SETUP\n\n")

	mode, err := d.askMode()
	if err != nil {
		return err
	}

	switch mode {
	case 1:
		err = d.createNewUser()

	case 2:
		err = d.joinNode()
	}

	return err
}

func (d *SetupDialogue) askMode() (int, error) {
	for {
		d.Say("1. Create a new user identity")
		d.Say("2. Join another node")

		i, err := d.AskInt("\nWhat would you like to do?")
		if err != nil {
			return 0, err
		}

		switch i {
		case 1, 2:
			return i, nil

		default:
			d.Say("Invalid choice.\n\n")
		}
	}
}

func (d *SetupDialogue) joinNode() error {
	return errors.New("not implemented")
}

func (d *SetupDialogue) createNewUser() error {
	alias, err := d.Ask("Choose an alias for the new user:")
	if err != nil {
		return err
	}

	d.Say("Creating a new identity...")
	identity, _, err := d.keys.CreateKey(alias)
	if err != nil {
		return err
	}

	d.Say("Setting local user identity...")
	err = d.user.SetLocalUser(identity)
	if err != nil {
		return err
	}

	d.Say("Generating access token...")
	token, err := d.apphost.CreateAccessToken(identity)
	if err != nil {
		return err
	}

	d.Say("YOUR ACCESS TOKEN IS: %s\n\n", token)
	d.Say("To use tools like anc from the command line, set ASTRALD_APPHOST_TOKEN environment variable like this:\n\n")
	d.Say("export ASTRALD_APPHOST_TOKEN=\"%s\"\n\n", token)

	add, err := d.AskBool("Add this line to your .profile file?")
	if err != nil {
		return err
	}

	if add {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		profilePath := path.Join(homeDir, ".profile")

		file, err := os.OpenFile(profilePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		defer file.Close()

		fmt.Fprintf(file, "\n# Added by astral\n")
		fmt.Fprintf(file, "export ASTRALD_APPHOST_TOKEN=\"%s\"\n", token)

		d.Say("Done. Changes to the .profile file will take effect on your next login.")
	}

	return nil
}
