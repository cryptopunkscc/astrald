package tc

import (
	"encoding/hex"
	"errors"
	"io/ioutil"
)

// Authenticate selects an auth method supported by both the daemon and this client and performs it.
// Returns an error if no mutually supported method is available.
func (ctl *Control) Authenticate() error {
	if ctl.ProtocolInfo().HasAuthMethod(authMethodCookie) {
		return ctl.authenticateWithCookie()
	}

	return errors.New("no supported auth method")
}

func (ctl *Control) authenticateWithCookie() error {
	bytes, err := ioutil.ReadFile(ctl.ProtocolInfo().AuthCookieFile)
	if err != nil {
		return err
	}

	cookie := hex.EncodeToString(bytes)

	_, _, err = ctl.request("AUTHENTICATE %s", cookie)

	return err
}
