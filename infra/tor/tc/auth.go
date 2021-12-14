package tc

import (
	"encoding/hex"
	"errors"
	"io/ioutil"
)

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
