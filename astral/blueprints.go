package astral

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"

	"github.com/cryptopunkscc/astrald/sig"
)

var DefaultBlueprints = &Blueprints{}

// Blueprints is a structure that holds prototypes of astral objects.
type Blueprints struct {
	Blueprints sig.Map[string, Object]
	Parent     *Blueprints
}

// NewBlueprints returns a new instance of Blueprints. If parent is not nil, it will be used by Make() to look up
// prototypes if not found in this instance.
func NewBlueprints(parent *Blueprints) *Blueprints {
	return &Blueprints{Parent: parent}
}

// HasBlueprints is used to check if a variable holds blueprints
type HasBlueprints interface {
	Blueprints() *Blueprints
}

// Make returns a zero-value instance of an object of the specified type
func (bp *Blueprints) Make(typeName string) Object {
	p, ok := bp.Blueprints.Get(typeName)
	if !ok {
		if bp.Parent != nil {
			return bp.Parent.Make(typeName)
		}
		return nil
	}
	var v = reflect.ValueOf(p)
	var c = reflect.New(v.Elem().Type())

	return c.Interface().(Object)
}

// Add adds a new object prototype
func (bp *Blueprints) Add(object ...Object) error {
	var errs []error

	for _, o := range object {
		if len(o.ObjectType()) == 0 {
			errs = append(errs, fmt.Errorf("object type is empty for %s", reflect.TypeOf(o)))
			continue
		}
		_, ok := bp.Blueprints.Set(o.ObjectType(), o)
		if !ok {
			errs = append(errs, fmt.Errorf("blueprint for %s already added", o.ObjectType()))
		}
	}

	return errors.Join(errs...)
}

// AddAs adds a new object prototype as provided type name, ignoring object's own ObjectType()
func (bp *Blueprints) AddAs(typeName string, object Object) error {
	_, ok := bp.Blueprints.Set(typeName, object)
	if !ok {
		return fmt.Errorf("blueprint for %s already added", typeName)
	}
	return nil
}

// Refine takes a RawObject and reparses it into a concrete object if a prototype for the type is found
func (bp *Blueprints) Refine(raw *RawObject) (Object, error) {
	b := bp.Make(raw.ObjectType())
	if b == nil {
		return nil, errors.New("blueprint not found")
	}

	var buf = &bytes.Buffer{}
	_, err := raw.WriteTo(buf)
	if err != nil {
		return nil, err
	}

	_, err = b.ReadFrom(buf)
	return b, err
}

// Names returns type names of all registered prototypes
func (bp *Blueprints) Names() (names []string) {
	if bp.Parent != nil {
		names = bp.Parent.Names()
	}
	return append(names, bp.Blueprints.Keys()...)
}

// Read reads an object from a reader and use blueprints to make an instance.
func (bp *Blueprints) Read(r io.Reader, canonical bool) (o Object, n int64, err error) {
	var typeName String8
	if canonical {
		var h ObjectHeader
		n, err = h.ReadFrom(r)
		typeName = String8(h)
	} else {
		n, err = typeName.ReadFrom(r)
	}
	if err != nil {
		return
	}
	if bp != nil {
		o = bp.Make(string(typeName))
	}
	if o == nil {
		o = &RawObject{Type: string(typeName)}
	}
	m, err := o.ReadFrom(&ReaderWithBlueprints{
		bp:     bp,
		Reader: r,
	})
	n += m
	return o, n, err
}

// Unpack unpacks a short form object from a buffer
func (bp *Blueprints) Unpack(data []byte) (object Object, err error) {
	object, _, err = bp.Read(bytes.NewReader(data), false)
	return
}

// Inject wraps an io.Reader in a wrapper that HasBlueprints
func (bp *Blueprints) Inject(r io.Reader) io.Reader {
	return NewReaderWithBlueprints(r, bp)
}

var _ HasBlueprints = &ReaderWithBlueprints{}
var _ io.Reader = &ReaderWithBlueprints{}

type ReaderWithBlueprints struct {
	bp *Blueprints
	io.Reader
}

func NewReaderWithBlueprints(reader io.Reader, bp *Blueprints) *ReaderWithBlueprints {
	return &ReaderWithBlueprints{Reader: reader, bp: bp}
}

func (r *ReaderWithBlueprints) Blueprints() *Blueprints {
	return r.bp
}

// ExtractBlueprints checks if the argument implements HasBlueprints. If yes, it returns its Blueprints(),
// otherwise it returns DefaultBlueprints.
func ExtractBlueprints(v any) (bp *Blueprints) {
	if s, ok := v.(HasBlueprints); ok {
		bp = s.Blueprints()
	}
	if bp == nil {
		bp = DefaultBlueprints
	}
	if bp == nil {
		bp = NewBlueprints(nil)
	}
	return
}
