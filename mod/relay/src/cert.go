package relay

import (
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/lib/adc"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/relay"
	"github.com/cryptopunkscc/astrald/object"
	"time"
)

func (mod *Module) indexData(objectID object.ID) error {
	// check if cert is already indexed
	if mod.isCertIndexed(objectID) {
		return relay.ErrCertAlreadyIndexed
	}

	r, err := mod.objects.Open(objectID, &objects.OpenOpts{Virtual: true})
	if err != nil {
		return err
	}
	defer r.Close()

	dataType, err := adc.ReadHeader(r)
	if err != nil {
		return err
	}

	switch dataType {
	case relay.CertType:
		var cert relay.Cert
		if err = cslq.Decode(r, "v", &cert); err != nil {
			return err
		}

		return mod.index(&cert)

	default:
		return nil
	}
}

func (mod *Module) Index(cert *relay.Cert) error {
	if err := cert.Verify(); err != nil {
		return err
	}

	return mod.index(cert)
}

func (mod *Module) index(cert *relay.Cert) error {
	var w = object.NewResolver()
	cslq.Encode(w, "vv", adc.Header(relay.CertType), cert)
	objectID := w.Resolve()

	return mod.db.Create(&dbCert{
		DataID:    objectID,
		Direction: string(cert.Direction),
		TargetID:  cert.TargetID,
		RelayID:   cert.RelayID,
		ExpiresAt: cert.ExpiresAt,
	}).Error
}

func (mod *Module) MakeCert(targetID id.Identity, relayID id.Identity, direction relay.Direction, duration time.Duration) (object.ID, error) {
	// create the certificate object
	cert, err := mod.makeCert(targetID, relayID, direction, duration)
	if err != nil {
		return object.ID{}, err
	}

	// create a data writer
	w, err := mod.objects.Create(nil)
	if err != nil {
		return object.ID{}, err
	}

	// write data type
	err = adc.WriteHeader(w, relay.CertType)
	if err != nil {
		return object.ID{}, err
	}

	// encode the certificate
	err = cslq.Encode(w, "v", cert)
	if err != nil {
		return object.ID{}, fmt.Errorf("encode error: %w", err)
	}

	// commit to storage
	objectID, err := w.Commit()
	if err != nil {
		return object.ID{}, err
	}

	// add the certificate to the index
	err = mod.indexData(objectID)
	if err != nil {
		return object.ID{}, err
	}

	return objectID, err
}

func (mod *Module) FindCerts(opts *relay.FindOpts) ([]object.ID, error) {
	var list []object.ID

	if opts == nil {
		opts = &relay.FindOpts{}
	}

	var q = mod.db.
		Model(&dbCert{}).
		Order("expires_at desc")

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

	return list, q.Select("data_id").Find(&list).Error
}

func (mod *Module) ReadCert(opts *relay.FindOpts) ([]byte, error) {
	certIDs, err := mod.FindCerts(opts)
	if err != nil {
		return nil, err
	}

	for _, certID := range certIDs {
		bytes, err := mod.objects.Get(certID, &objects.OpenOpts{Virtual: true})
		if err != nil {
			mod.log.Errorv(2, "error reading %v: %v", certID, err)
			continue
		}
		return bytes, nil
	}

	return nil, relay.ErrCertNotFound
}

func (mod *Module) LoadCert(objectID object.ID) (*relay.Cert, error) {
	r, err := mod.objects.Open(objectID, &objects.OpenOpts{Virtual: true})
	if err != nil {
		return nil, err
	}
	defer r.Close()

	dataType, err := adc.ReadHeader(r)
	if err != nil {
		return nil, err
	}

	switch dataType {
	case relay.CertType:
		var cert relay.Cert
		if err = cslq.Decode(r, "v", &cert); err != nil {
			return nil, err
		}

		return &cert, nil
	}

	return nil, errors.New("invalid data type")
}

func (mod *Module) FindExternalRelays(targetID id.Identity) ([]id.Identity, error) {
	var relays []id.Identity

	var q = mod.db.
		Model(&dbCert{}).
		Where("relay_id != ? and target_id = ? and expires_at > ? and direction in ?",
			mod.node.Identity(),
			targetID,
			time.Now(),
			[]relay.Direction{relay.Inbound, relay.Both},
		)

	var err = q.
		Select("relay_id").
		Find(&relays).
		Error
	if err != nil {
		return nil, err
	}

	return relays, nil
}

func (mod *Module) makeCert(
	targetID id.Identity,
	relayID id.Identity,
	direction relay.Direction,
	duration time.Duration,
) (*relay.Cert, error) {
	var err error

	if duration == 0 {
		duration = relay.DefaultCertDuration
	}

	var cert = relay.Cert{
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

func (mod *Module) isCertIndexed(objectID object.ID) bool {
	var c int64
	var tx = mod.db.Model(&dbCert{}).Where("data_id = ?", objectID.String()).Count(&c)
	if tx.Error != nil {
		mod.log.Errorv(1, "database error: %v", tx.Error)
	}
	return c > 0
}
