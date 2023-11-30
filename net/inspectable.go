package net

import "sync"

type SourceGetter interface {
	Source() any
}

type SourceSetter interface {
	SetSource(any) error
}

type SourceGetSetter interface {
	SourceGetter
	SourceSetter
}

var _ SourceGetSetter = &SourceField{}

type SourceField struct {
	source any
}

func NewSourceField(source any) *SourceField {
	return &SourceField{source: source}
}

func (field *SourceField) Source() any {
	return field.source
}

func (field *SourceField) SetSource(source any) error {
	field.source = source
	return nil
}

type OutputGetter interface {
	Output() SecureWriteCloser
}

type OutputSetter interface {
	SetOutput(SecureWriteCloser) error
}

type OutputGetSetter interface {
	OutputGetter
	OutputSetter
}

var _ OutputGetSetter = &OutputField{}

type OutputField struct {
	parent any
	output SecureWriteCloser
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

	return field.output
}

func (field *OutputField) SetOutput(output SecureWriteCloser) error {
	field.Lock()
	defer field.Unlock()

	if s, ok := field.output.(SourceGetSetter); ok {
		if s.Source() == field.parent {
			s.SetSource(nil)
		}
	}
	field.output = output
	if s, ok := output.(SourceGetSetter); ok {
		if s.Source() == nil {
			s.SetSource(field.parent)
		}
	}
	return nil
}

func RootSource(src any) any {
	if src == nil {
		return nil
	}

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
