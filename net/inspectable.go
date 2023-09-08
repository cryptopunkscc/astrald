package net

import "sync"

type SourceGetter interface {
	Source() any
}

type SourceSetter interface {
	SetSource(any) error
}

var _ SourceGetter = &SourceField{}
var _ SourceSetter = &SourceField{}

type SourceField struct {
	source any
}

func NewSourceField(source any) *SourceField {
	return &SourceField{source: source}
}

func (s *SourceField) Source() any {
	return s.source
}

func (s *SourceField) SetSource(source any) error {
	s.source = source
	return nil
}

type OutputGetter interface {
	Output() SecureWriteCloser
}

type OutputSetter interface {
	SetOutput(SecureWriteCloser) error
}

type OutputField struct {
	parent any
	SecureWriteCloser
	sync.Mutex
}

func NewOutputField(parent any, writer SecureWriteCloser) *OutputField {
	f := &OutputField{parent: parent}
	f.SetOutput(writer)
	return f
}

func (field *OutputField) Output() SecureWriteCloser {
	field.Lock()
	defer field.Unlock()

	return field.SecureWriteCloser
}

func (field *OutputField) SetOutput(output SecureWriteCloser) error {
	field.Lock()
	defer field.Unlock()

	if s, ok := field.SecureWriteCloser.(SourceSetter); ok {
		s.SetSource(nil)
	}
	field.SecureWriteCloser = output
	if s, ok := output.(SourceSetter); ok {
		s.SetSource(field.parent)
	}
	return nil
}

func RootSource(src any) any {
	for {
		getter, ok := src.(SourceGetter)
		if !ok {
			break
		}
		if getter.Source() == nil {
			break
		}
		src = getter.Source()
	}
	return src
}

func FinalOutput(output any) SecureWriteCloser {
	final, ok := output.(SecureWriteCloser)
	if !ok {
		if getter, ok := output.(OutputGetter); ok {
			final = getter.Output()
		}
	}

	for {
		getter, ok := final.(OutputGetter)
		if !ok {
			break
		}
		if getter.Output() == nil {
			break
		}
		final = getter.Output()
	}
	return final
}
