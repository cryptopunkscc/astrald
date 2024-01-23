package relay

import (
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/lib/adc"
	"github.com/cryptopunkscc/astrald/mod/relay"
	"time"
)

const defaultCertDuration = 100 * 365 * 24 * time.Hour

func (mod *Module) IndexCert(dataID data.ID) error {
	// check if cert is already indexed
	if mod.isCertIndexed(dataID) {
		return relay.ErrCertAlreadyIndexed
	}

	r, err := mod.storage.Read(dataID, nil)
	if err != nil {
		return err
	}
	defer r.Close()

	dataType, err := adc.ReadHeader(r)
	if err != nil {
		return err
	}

	switch dataType {
	case relay.RelayCertType:
		var cert relay.RelayCert
		if err = cslq.Decode(r, "v", &cert); err != nil {
			return err
		}
		if err := cert.Verify(); err != nil {
			return err
		}
		tx := mod.db.Create(&dbRelayCert{
			DataID:    dataID.String(),
			Direction: string(cert.Direction),
			TargetID:  cert.TargetID.PublicKeyHex(),
			RelayID:   cert.RelayID.PublicKeyHex(),
			ExpiresAt: cert.ExpiresAt,
		})
		return tx.Error

	default:
		return nil
	}
}

func (mod *Module) MakeCert(targetID id.Identity, relayID id.Identity, direction relay.Direction, duration time.Duration) (data.ID, error) {
	// create the certificate object
	cert, err := mod.makeCert(targetID, relayID, direction, duration)
	if err != nil {
		return data.ID{}, err
	}

	// create a data writer
	w, err := mod.storage.Store(nil)
	if err != nil {
		return data.ID{}, err
	}

	// write data type
	err = adc.WriteHeader(w, relay.RelayCertType)
	if err != nil {
		return data.ID{}, err
	}

	// encode the certificate
	err = cslq.Encode(w, "v", cert)
	if err != nil {
		return data.ID{}, fmt.Errorf("encode error: %w", err)
	}

	// commit to storage
	dataID, err := w.Commit()
	if err != nil {
		return data.ID{}, err
	}

	// add the certificate to the index
	err = mod.IndexCert(dataID)
	if err != nil {
		return data.ID{}, err
	}

	return dataID, err
}

func (mod *Module) FindCerts(opts *relay.FindOpts) ([]data.ID, error) {
	var rows []dbRelayCert

	if opts == nil {
		opts = &relay.FindOpts{}
	}

	var q = mod.db

	if !opts.RelayID.IsZero() {
		q = q.Where("relay_id = ?", opts.RelayID.PublicKeyHex())
	}

	if !opts.ExcludeRelayID.IsZero() {
		q = q.Where("relay_id != ?", opts.ExcludeRelayID.PublicKeyHex())
	}

	if !opts.TargetID.IsZero() {
		q = q.Where("target_id = ?", opts.TargetID.PublicKeyHex())
	}

	if !opts.ExcludeTargetID.IsZero() {
		q = q.Where("target_id != ?", opts.ExcludeTargetID.PublicKeyHex())
	}

	if opts.Direction != "" {
		switch opts.Direction {
		case relay.Inbound:
			q = q.Where("direction in ?", []relay.Direction{relay.Inbound, relay.Both})
		case relay.Outbound:
			q = q.Where("direction in ?", []relay.Direction{relay.Outbound, relay.Both})
		case relay.Both:
			q = q.Where("direction = ?", relay.Both)
		}
	}

	if !opts.IncludeExpired {
		q = q.Where("expires_at > ?", time.Now())
	}

	var tx = q.Order("expires_at desc").Find(&rows)
	if tx.Error != nil {
		return nil, tx.Error
	}

	var list []data.ID

	for _, row := range rows {
		dataID, err := data.Parse(row.DataID)
		if err != nil {
			mod.log.Errorv(2, "error parsing %v: %v", row.DataID, err)
			continue
		}
		list = append(list, dataID)
	}

	return list, nil
}

func (mod *Module) ReadCert(opts *relay.FindOpts) ([]byte, error) {
	certIDs, err := mod.FindCerts(opts)
	if err != nil {
		return nil, err
	}

	for _, certID := range certIDs {
		bytes, err := mod.storage.ReadAll(certID, nil)
		if err != nil {
			mod.log.Errorv(2, "error reading %v: %v", certID, err)
			continue
		}
		return bytes, nil
	}

	return nil, relay.ErrCertNotFound
}

func (mod *Module) LoadCert(dataID data.ID) (*relay.RelayCert, error) {
	r, err := mod.storage.Read(dataID, nil)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	dataType, err := adc.ReadHeader(r)
	if err != nil {
		return nil, err
	}

	switch dataType {
	case relay.RelayCertType:
		var cert relay.RelayCert
		if err = cslq.Decode(r, "v", &cert); err != nil {
			return nil, err
		}

		return &cert, nil
	}

	return nil, errors.New("invalid data type")
}

func (mod *Module) FindExternalRelays(targetID id.Identity) ([]id.Identity, error) {
	var rows []dbRelayCert
	var tx = mod.db.Where("relay_id != ? and target_id = ? and expires_at > ? and direction in ?",
		mod.node.Identity().PublicKeyHex(),
		targetID.PublicKeyHex(),
		time.Now(),
		[]relay.Direction{relay.Inbound, relay.Both},
	).Find(&rows)

	if tx.Error != nil {
		return nil, tx.Error
	}

	var relays []id.Identity
	for _, row := range rows {
		relayID, err := id.ParsePublicKeyHex(row.RelayID)
		if err != nil {
			continue
		}
		relays = append(relays, relayID)
	}

	return relays, nil
}

func (mod *Module) makeCert(
	targetID id.Identity,
	relayID id.Identity,
	direction relay.Direction,
	duration time.Duration,
) (*relay.RelayCert, error) {
	var err error

	if duration == 0 {
		duration = defaultCertDuration
	}

	var cert = relay.RelayCert{
		TargetID:  targetID,
		RelayID:   relayID,
		Direction: direction,
		ExpiresAt: time.Now().Add(duration),
	}

	// sign with target identity
	cert.TargetSig, err = mod.keys.Sign(cert.TargetID, cert.Hash())
	if err != nil {
		return nil, fmt.Errorf("error signing certificate with target key: %w", err)
	}

	// sign with relay identity
	cert.RelaySig, err = mod.keys.Sign(relayID, cert.Hash())
	if err != nil {
		return nil, fmt.Errorf("error signing certificate with relay key: %w", err)
	}

	// check certificate's validity
	if err = cert.Validate(); err != nil {
		return nil, fmt.Errorf("generated certificate is invalid: %w", err)
	}

	return &cert, nil
}

func (mod *Module) isCertIndexed(dataID data.ID) bool {
	var c int64
	var tx = mod.db.Model(&dbRelayCert{}).Where("data_id = ?", dataID.String()).Count(&c)
	if tx.Error != nil {
		mod.log.Errorv(1, "database error: %v", tx.Error)
	}
	return c > 0
}
