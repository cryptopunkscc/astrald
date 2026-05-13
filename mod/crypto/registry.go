package crypto

// Registry maps (keyType, scheme) pairs to capability handlers.
// It provides O(1) dispatch — no iteration or error-based negotiation.
type Registry struct {
	keyDerivers   map[string]KeyDeriver
	hashVerifiers map[string]map[string]HashVerifier
	textVerifiers map[string]map[string]TextVerifier
	hashFactories map[string]map[string]HashSignerFactory
	textFactories map[string]map[string]TextSignerFactory
}

func NewRegistry() *Registry {
	return &Registry{
		keyDerivers:   make(map[string]KeyDeriver),
		hashVerifiers: make(map[string]map[string]HashVerifier),
		textVerifiers: make(map[string]map[string]TextVerifier),
		hashFactories: make(map[string]map[string]HashSignerFactory),
		textFactories: make(map[string]map[string]TextSignerFactory),
	}
}

// RegisterKeyDeriver registers a handler for deriving public keys from private keys.
func (r *Registry) RegisterKeyDeriver(keyType string, deriver KeyDeriver) {
	r.keyDerivers[keyType] = deriver
}

// RegisterHashVerifier registers a handler for verifying hash signatures.
func (r *Registry) RegisterHashVerifier(keyType, scheme string, verifier HashVerifier) {
	if r.hashVerifiers[keyType] == nil {
		r.hashVerifiers[keyType] = make(map[string]HashVerifier)
	}
	r.hashVerifiers[keyType][scheme] = verifier
}

// RegisterTextVerifier registers a handler for verifying text signatures.
func (r *Registry) RegisterTextVerifier(keyType, scheme string, verifier TextVerifier) {
	if r.textVerifiers[keyType] == nil {
		r.textVerifiers[keyType] = make(map[string]TextVerifier)
	}
	r.textVerifiers[keyType][scheme] = verifier
}

// RegisterHashSignerFactory registers a factory for creating hash signers.
func (r *Registry) RegisterHashSignerFactory(keyType, scheme string, factory HashSignerFactory) {
	if r.hashFactories[keyType] == nil {
		r.hashFactories[keyType] = make(map[string]HashSignerFactory)
	}
	r.hashFactories[keyType][scheme] = factory
}

// RegisterTextSignerFactory registers a factory for creating text signers.
func (r *Registry) RegisterTextSignerFactory(keyType, scheme string, factory TextSignerFactory) {
	if r.textFactories[keyType] == nil {
		r.textFactories[keyType] = make(map[string]TextSignerFactory)
	}
	r.textFactories[keyType][scheme] = factory
}

// Lookup methods — return nil if unsupported.

func (r *Registry) LookupKeyDeriver(keyType string) KeyDeriver {
	return r.keyDerivers[keyType]
}

func (r *Registry) LookupHashVerifier(keyType, scheme string) HashVerifier {
	if m := r.hashVerifiers[keyType]; m != nil {
		return m[scheme]
	}
	return nil
}

func (r *Registry) LookupTextVerifier(keyType, scheme string) TextVerifier {
	if m := r.textVerifiers[keyType]; m != nil {
		return m[scheme]
	}
	return nil
}

func (r *Registry) LookupHashSignerFactory(keyType, scheme string) HashSignerFactory {
	if m := r.hashFactories[keyType]; m != nil {
		return m[scheme]
	}
	return nil
}

func (r *Registry) LookupTextSignerFactory(keyType, scheme string) TextSignerFactory {
	if m := r.textFactories[keyType]; m != nil {
		return m[scheme]
	}
	return nil
}
