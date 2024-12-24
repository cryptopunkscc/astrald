package astral

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/cryptopunkscc/astrald/object"
	"io"
	"slices"
	"sync"
)

var _ Object = &Bundle{}

// Bundle is an ordered collection of unique objects of various types.
type Bundle struct {
	objects []Object
	index   map[string]int
	mu      sync.Mutex
}

func (*Bundle) ObjectType() string { return "astral.bundle" }

// NewBundle returns a new Bundle instance. objects can be nil.
func NewBundle() *Bundle {
	return &Bundle{}
}

// Append appends objects to the bundle.
func (b *Bundle) Append(objects ...Object) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.index == nil {
		b.index = make(map[string]int)
	}

	for _, o := range objects {
		err := b.append(o)
		if err != nil {
			return err
		}
	}

	return nil
}

// Fetch fetches the object from the Bundle. Returns nil if the object is not in the Bundle.
func (b *Bundle) Fetch(objectID object.ID) Object {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.index == nil {
		return nil
	}

	idstr := objectID.String()
	idx, ok := b.index[idstr]
	if !ok {
		return nil
	}

	return b.objects[idx]
}

// WriteTo writes Bundle's payload to the writer.
func (b Bundle) WriteTo(w io.Writer) (n int64, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// write object count
	n, err = Uint32(len(b.objects)).WriteTo(w)
	if err != nil {
		return
	}

	// write objects
	var m int64
	for _, o := range b.objects {
		var buf = &bytes.Buffer{}

		// write the type to the buffer
		_, err = String8(o.ObjectType()).WriteTo(buf)
		if err != nil {
			return
		}

		// write the payload to the buffer
		_, err = o.WriteTo(buf)
		if err != nil {
			return
		}

		// write the length-encoded buffer to the writer
		m, err = Bytes32(buf.Bytes()).WriteTo(w)
		n += m
		if err != nil {
			return
		}
	}

	return
}

// ReadFrom reads Bundle's payload from the reader.
func (b *Bundle) ReadFrom(r io.Reader) (n int64, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.objects = nil
	b.index = make(map[string]int)

	// read object count
	var count Uint32
	n, err = count.ReadFrom(r)
	if err != nil {
		return
	}

	var m int64
	var o Object
	for i := 0; i < int(count); i++ {
		// read a 32-bit length-encoded buffer
		var buf = Bytes32{}
		m, err = buf.ReadFrom(r)
		n += m
		if err != nil {
			return
		}

		// read the object in the buffer
		o, _, err = ExtractBlueprints(r).Read(bytes.NewReader(buf), false)
		if err != nil {
			return
		}

		err = b.append(o)
		if err != nil {
			return
		}
	}

	return
}

// Objects returns a copy of the underlying list of objects
func (b *Bundle) Objects() []Object {
	b.mu.Lock()
	defer b.mu.Unlock()

	return slices.Clone(b.objects)
}

func (b *Bundle) String() string {
	return fmt.Sprintf("bundle (%d objects)", len(b.objects))
}

func (b *Bundle) MarshalJSON() ([]byte, error) {
	type j struct {
		Type   string
		Object Object
	}

	var list []j
	for _, o := range b.objects {
		list = append(list, j{
			Type:   o.ObjectType(),
			Object: o,
		})
	}

	return json.Marshal(list)
}

func (b *Bundle) append(object Object) error {
	objectID, err := ResolveObjectID(object)
	if err != nil {
		return fmt.Errorf("error resolving object id: %w", err)
	}

	// skip duplicates
	idstr := objectID.String()
	if _, found := b.index[idstr]; found {
		return fmt.Errorf("duplicate object")
	}

	b.objects = append(b.objects, object)
	b.index[idstr] = len(b.objects) - 1
	return nil
}

// SelectByType select objects of the parameter type from a generic list of Objects
func SelectByType[T Object](objects []Object) (list []T) {
	for _, o := range objects {
		if o, ok := o.(T); ok {
			list = append(list, o)
		}
	}
	return
}

// First returns the first object of the parameter type from a generic list of Objects
func First[T Object](objects []Object) (object T, found bool) {
	for _, o := range objects {
		if o, ok := o.(T); ok {
			return o, true
		}
	}
	return
}

// Fetch fetches a type-cast object from the bundle
func Fetch[T Object](bundle *Bundle, objectID object.ID) (object T, found bool) {
	object, found = bundle.Fetch(objectID).(T)
	return
}

func init() {
	DefaultBlueprints.Add(&Bundle{})
}
