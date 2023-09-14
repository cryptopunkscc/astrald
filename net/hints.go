package net

type Hints struct {
	Origin        string
	AllowRedirect bool
	DontMonitor   bool
	Extra         map[string]any
	_             struct{}
}

func DefaultHints() Hints {
	return Hints{Origin: OriginLocal}
}

func (hints Hints) SetAllowRedirect() Hints {
	var clone = hints.clone()
	clone.AllowRedirect = true
	return clone
}

func (hints Hints) SetDontMonitor() Hints {
	var clone = hints.clone()
	clone.DontMonitor = true
	return clone
}

func (hints Hints) WithOrigin(origin string) Hints {
	var clone = hints.clone()
	clone.Origin = origin
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
