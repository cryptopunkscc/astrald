package astral

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"

	"github.com/cryptopunkscc/astrald/sig"
)

var DefaultBlueprints = &Blueprints{
	TypeReader: ReadShortType,
	TypeWriter: WriteShortType,
}

// Blueprints is a structure that holds prototypes of astral objects.
type Blueprints struct {
	Blueprints sig.Map[string, Object]
	Parent     *Blueprints
	TypeReader TypeReader
	TypeWriter TypeWriter
}

// HasBlueprints is used to check if a variable holds blueprints
type HasBlueprints interface {
	Blueprints() *Blueprints
}

// NewBlueprints returns a new instance of Blueprints. If parent is not nil, it will be used by New() to look up
// prototypes if not found in this instance.
func NewBlueprints(parent *Blueprints) *Blueprints {
	return &Blueprints{
		Parent:     parent,
		TypeReader: ReadShortType,
		TypeWriter: WriteShortType,
	}
}

// Canonical returns a child Blueprints object that uses canonical type encoding
func (bp *Blueprints) Canonical() *Blueprints {
	return &Blueprints{
		Parent:     bp,
		TypeReader: ReadCanonicalType,
		TypeWriter: WriteCanonicalType,
	}
}

// Short returns a child Blueprints object that uses short type encoding
func (bp *Blueprints) Short() *Blueprints {
	return &Blueprints{
		Parent:     bp,
		TypeReader: ReadShortType,
		TypeWriter: WriteShortType,
	}
}

// Indexed returns a child Blueprints object that encodes types as uint8 indicating the index of the type
func (bp *Blueprints) Indexed(types []string) *Blueprints {
	var rev = map[string]Uint8{}
	for idx, v := range types {
		rev[v] = Uint8(idx)
	}

	rbp := &Blueprints{Parent: bp}
	rbp.TypeReader = func(r io.Reader) (t ObjectType, n int64, err error) {
		var code Uint8
		n, err = code.ReadFrom(r)
		if err != nil {
			return
		}

		if int(code) >= len(types) {
			err = fmt.Errorf("invalid type code %d", code)
			return
		}

		return ObjectType(types[code]), n, nil
	}

	rbp.TypeWriter = func(w io.Writer, t ObjectType) (n int64, err error) {
		var code Uint8
		var ok bool

		code, ok = rev[string(t)]
		if !ok {
			return 0, errors.New("invalid type")
		}

		return code.WriteTo(w)
	}

	return rbp
}

// New returns a zero-value object of the specified type or nil if no blueprint is found.
func (bp *Blueprints) New(typeName string) Object {
	p, ok := bp.Blueprints.Get(typeName)
	if !ok {
		if bp.Parent != nil {
			return bp.Parent.New(typeName)
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

// Types returns type names of all registered object types
func (bp *Blueprints) Types() (names []string) {
	if bp.Parent != nil {
		names = bp.Parent.Types()
	}
	return append(names, bp.Blueprints.Keys()...)
}

// Read reads an object from the reader.
func (bp *Blueprints) Read(r io.Reader) (o Object, n int64, err error) {
	return bp.read(r, bp.TypeReader)
}

// read reads an object from the reader using the provided type reader to read the type.
func (bp *Blueprints) read(r io.Reader, readType TypeReader) (object Object, n int64, err error) {
	// read the object type
	var objectType ObjectType
	var m int64

	objectType, n, err = readType(r)
	if err != nil {
		return
	}

	if len(objectType) == 0 {
		return nil, 0, errors.New("empty object type")
	}

	// make a new object of the type
	object = bp.New(string(objectType))
	if object == nil {
		return nil, 0, newErrBlueprintNotFound(string(objectType))
	}

	// inject blueprints into the reader with default (short) type encoding
	r = &ReaderWithBlueprints{
		bp:     bp.Short(),
		Reader: r,
	}

	// read the object payload
	m, err = object.ReadFrom(r)
	n += m
	return object, n, err
}

// Write writes an object to the writer.
func (bp *Blueprints) Write(w io.Writer, object Object) (n int64, err error) {
	return bp.write(w, object, bp.TypeWriter)
}

// write writes an object to the writer using the provided type writer to write the type.
func (bp *Blueprints) write(w io.Writer, object Object, write TypeWriter) (n int64, err error) {
	var m int64

	n, err = write(w, ObjectType(object.ObjectType()))
	if err != nil {
		return
	}

	m, err = object.WriteTo(w)
	n += m
	return
}

// Unpack unpacks an object from a buffer
func (bp *Blueprints) Unpack(data []byte) (object Object, err error) {
	object, _, err = bp.Read(bytes.NewReader(data))
	return
}

// Pack packs an object into a buffer
func (bp *Blueprints) Pack(object Object) (data []byte, err error) {
	var buf = &bytes.Buffer{}
	_, err = bp.Write(buf, object)
	return buf.Bytes(), err
}

// Inject wraps an io.Reader in a wrapper that HasBlueprints
func (bp *Blueprints) Inject(r io.Reader) io.Reader {
	return NewReaderWithBlueprints(r, bp)
}

// Unpack unpacks an object from the buffer using DefaultBlueprints.

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

// Write writes the object to the writer using DefaultBlueprints
func Write(w io.Writer, obj Object) (_ int64, err error) {
	return DefaultBlueprints.Write(w, obj)
}

// Pack writes the object in its short form to a buffer and returns the buffer
func Pack(obj Object) ([]byte, error) {
	return DefaultBlueprints.Pack(obj)
}

// Add adds the object prototypes to the default Blueprints
func Add(object ...Object) error {
	return DefaultBlueprints.Add(object...)
}
