package file

import (
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/storage/file"
	"github.com/cryptopunkscc/astrald/components/uid"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"path"
)

const dirName = "identities"

type identities struct {
	dir string
}

func NewIdentities(home string) uid.Identities {
	storageDir, err := file.ResolveDir(home, dirName)
	if err != nil {
		panic(err)
	}
	return &identities{
		dir: storageDir,
	}
}

func (i *identities) Update(card uid.Card) error {
	fp := path.Join(i.dir, string(card.Id))
	open, err := os.OpenFile(fp, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	bytes, err := yaml.Marshal(card)
	if err != nil {
		return err
	}
	_, err = open.Write(bytes)
	if err != nil {
		return err
	}
	err = open.Sync()
	if err != nil {
		return err
	}
	return nil
}

func (i *identities) List() ([]uid.Card, error) {
	names, err := file.ListNames(i.dir)
	if err != nil {
		return nil, err
	}
	cards := make([]uid.Card, len(names))
	for index, name := range names {
		card, _ := i.Get(api.Identity(name))
		cards[index] = *card
	}
	return cards, nil
}

func (i *identities) Get(id api.Identity) (card *uid.Card, err error) {
	fp := path.Join(i.dir, string(id))
	f, err := os.Open(fp)
	if err != nil {
		return nil, nil
	}
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, nil
	}
	card = &uid.Card{}
	err = yaml.Unmarshal(b, card)
	if err != nil {
		return nil, nil
	}
	return card, nil
}
