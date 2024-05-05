package user

import (
	"bytes"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/lib/adc"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/relay"
)

func (mod *Module) loadCert(userID id.Identity, nodeID id.Identity, autoCreate bool) (cert []byte, err error) {
	// get user certificate for local relay
	cert, err = mod.relay.ReadCert(&relay.FindOpts{
		TargetID:  userID,
		RelayID:   nodeID,
		Direction: relay.Both,
	})

	if autoCreate && errors.Is(err, relay.ErrCertNotFound) {
		return mod.makeCert(userID, nodeID)
	}

	return
}

func (mod *Module) makeCert(userID id.Identity, nodeID id.Identity) (cert []byte, err error) {
	mod.log.Info("generating certificate for %v@%v...", userID, nodeID)
	certID, err := mod.relay.MakeCert(userID, nodeID, relay.Both, 0)
	if err != nil {
		mod.log.Info("error generating certificate for %v@%v: %v", userID, nodeID, err)
		return
	}

	return mod.objects.Get(certID, objects.DefaultOpenOpts)
}

func (mod *Module) checkCert(relayID id.Identity, certBytes []byte) error {
	var r = bytes.NewReader(certBytes)

	var dataType adc.Header
	err := cslq.Decode(r, "v", &dataType)
	if err != nil {
		return err
	}
	if dataType != relay.CertType {
		return errors.New("invalid data type")
	}

	var cert relay.Cert

	err = cslq.Decode(r, "v", &cert)
	if err != nil {
		return err
	}

	err = cert.Validate()
	if err != nil {
		return err
	}

	if !cert.RelayID.IsEqual(relayID) {
		mod.log.Errorv(2, "%v is not %v", cert.RelayID, relayID)
		return errors.New("relay mismatch")
	}

	mod.objects.Put(certBytes, nil)

	return nil
}
