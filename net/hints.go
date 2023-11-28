package net

type Hints struct {
	Origin  string // Origin denotes the where the query originated (OriginLocal or OriginNetwork)
	Silent  bool   // Silent tells the router to not log the query
	Reroute bool   // Reroute allows a nonce to reenter the router event though it's already en route
	Update  bool   // Update tells the monitored router to update query details for the nonce when rerouting
	Extra   map[string]any
	_       struct{}
}

func DefaultHints() Hints {
	return Hints{Origin: OriginLocal}
}

func (hints Hints) WithOrigin(origin string) Hints {
	var clone = hints.clone()
	clone.Origin = origin
	return clone
}

func (hints Hints) SetSilent() Hints {
	var clone = hints.clone()
	clone.Silent = true
	return clone
}

func (hints Hints) SetReroute() Hints {
	var clone = hints.clone()
	clone.Reroute = true
	return clone
}

func (hints Hints) SetUpdate() Hints {
	var clone = hints.clone()
	clone.Update = true
	return clone
}

func (hints Hints) WithValue(key string, val any) Hints {
	var clone = hints.clone()
	if clone.Extra == nil {
		clone.Extra = map[string]any{key: val}
	} else {
		clone.Extra[key] = val
	}
	return clone
}

func (hints Hints) Value(key string) (any, bool) {
	if hints.Extra != nil {
		val, ok := hints.Extra[key]
		return val, ok
	}
	return nil, false
}

func (hints Hints) clone() Hints {
	var clone = hints
	clone.Extra = map[string]any{}
	for k, v := range hints.Extra {
		clone.Extra[k] = v
	}
	return clone
}
