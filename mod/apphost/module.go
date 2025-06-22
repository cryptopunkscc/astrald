package apphost

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"io"
	"strings"
)

const ModuleName = "apphost"
const DBPrefix = "apphost__"

type Module interface {
	CreateAccessToken(*astral.Identity, astral.Duration) (*AccessToken, error)
	ActiveLocalAppContracts() ([]*AppContract, error)
}

type AccessToken struct {
	Identity  *astral.Identity
	Token     astral.String8
	ExpiresAt astral.Time
}

// astral

var _ astral.Object = &AccessToken{}

func (at AccessToken) ObjectType() string { return "astrald.mod.apphost.access_token" }

func (at AccessToken) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(at).WriteTo(w)
}

func (at *AccessToken) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(at).ReadFrom(r)
}

// json

func (at AccessToken) MarshalJSON() ([]byte, error) {
	type alias AccessToken
	return json.Marshal(alias(at))
}

func (at *AccessToken) UnmarshalJSON(bytes []byte) error {
	type alias AccessToken
	var a alias

	err := json.Unmarshal(bytes, &a)
	if err != nil {
		return err
	}

	*at = AccessToken(a)
	return nil
}

// text

func (at AccessToken) MarshalText() (text []byte, err error) {
	s := fmt.Sprintf("%s,%s,%s", at.Identity, at.Token, at.ExpiresAt)
	return []byte(s), nil
}

func (at *AccessToken) UnmarshalText(text []byte) (err error) {
	parts := strings.SplitN(string(text), ",", 3)
	if len(parts) != 3 {
		return errors.New("invalid format")
	}

	id := &astral.Identity{}
	err = id.UnmarshalText([]byte(parts[0]))
	if err != nil {
		return err
	}
	at.Identity = id
	at.Token = astral.String8(parts[1])

	var t astral.Time
	err = t.UnmarshalText([]byte(parts[2]))
	if err != nil {
		return err
	}
	at.ExpiresAt = t

	return nil
}

func init() {
	_ = astral.DefaultBlueprints.Add(&AccessToken{})
}
