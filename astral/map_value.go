package astral

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strconv"
)

type mapValue struct {
	reflect.Value
}

var _ Object = &mapValue{}

// astral:blueprint-ignore
func (val mapValue) ObjectType() string {
	return ""
}

// WriteTo emits the canonical map wire format:
//
//	[uint32 count] [encoded_key | encoded_value]…   sorted by encoded_key
//
// Keys must be reflect.String (encoded as String16) or unsigned integers of width 1/2/4/8
// (encoded as fixed-width big-endian). Values are encoded via objectify on their static type,
// which naturally produces tag+payload for interface element types and bare payload for
// concrete Object types — matching StringMap/IntMap heterogeneous and homogeneous modes.
func (val mapValue) WriteTo(w io.Writer) (n int64, err error) {
	keyWidth, ok := supportedMapKey(val.Type().Key().Kind())
	if !ok {
		return 0, fmt.Errorf("map_value: unsupported key kind %s", val.Type().Key().Kind())
	}

	err = binary.Write(w, ByteOrder, uint32(val.Len()))
	if err != nil {
		return
	}
	n += 4

	type pair struct{ key, value []byte }
	pairs := make([]pair, 0, val.Len())
	flagged := elemNeedsPresenceFlag(val.Type().Elem())

	for _, k := range val.MapKeys() {
		var keyBuf, valBuf bytes.Buffer

		err = writeMapKey(&keyBuf, k, keyWidth)
		if err != nil {
			return
		}

		o, oerr := objectify(addressableMapValue(val.MapIndex(k)))
		if oerr != nil {
			err = oerr
			return
		}
		if flagged {
			valBuf.Write(presenceFlagOne)
		}
		if _, err = o.WriteTo(&valBuf); err != nil {
			return
		}

		pairs = append(pairs, pair{key: keyBuf.Bytes(), value: valBuf.Bytes()})
	}

	sort.Slice(pairs, func(i, j int) bool {
		return bytes.Compare(pairs[i].key, pairs[j].key) < 0
	})

	for _, p := range pairs {
		var written int
		written, err = w.Write(p.key)
		n += int64(written)
		if err != nil {
			return
		}
		written, err = w.Write(p.value)
		n += int64(written)
		if err != nil {
			return
		}
	}

	return
}

func (val mapValue) ReadFrom(r io.Reader) (n int64, err error) {
	keyWidth, ok := supportedMapKey(val.Type().Key().Kind())
	if !ok {
		return 0, fmt.Errorf("map_value: unsupported key kind %s", val.Type().Key().Kind())
	}

	var l uint32
	err = binary.Read(r, ByteOrder, &l)
	if err != nil {
		return
	}
	n += 4

	if l == 0 {
		val.SetZero()
		return
	}

	val.Set(reflect.MakeMapWithSize(val.Type(), int(l)))
	flagged := elemNeedsPresenceFlag(val.Type().Elem())

	for range l {
		var m int64

		key := reflect.New(val.Type().Key()).Elem()
		m, err = readMapKey(r, key, keyWidth)
		n += m
		if err != nil {
			return
		}

		value := reflect.New(val.Type().Elem()).Elem()
		o, oerr := objectify(value)
		if oerr != nil {
			err = oerr
			return
		}
		if flagged {
			m, err = consumePresenceFlag(r)
			n += m
			if err != nil {
				return
			}
		}
		m, err = o.ReadFrom(r)
		n += m
		if err != nil {
			return
		}

		val.SetMapIndex(key, value)
	}

	return
}

// supportedMapKey reports the canonical key width for an allowed map-key kind.
// Width 0 means String16 wire (reflect.String). Widths 1/2/4/8 are fixed-width big-endian
// unsigned integers. reflect.Uint is rejected: its width is platform-dependent, which would
// make the wire bytes non-portable and break content-addressing.
func supportedMapKey(k reflect.Kind) (width uint8, ok bool) {
	switch k {
	case reflect.String:
		return 0, true
	case reflect.Uint8:
		return 1, true
	case reflect.Uint16:
		return 2, true
	case reflect.Uint32:
		return 4, true
	case reflect.Uint64:
		return 8, true
	}
	return 0, false
}

func writeMapKey(w io.Writer, k reflect.Value, width uint8) error {
	if width == 0 {
		_, err := String16(k.String()).WriteTo(w)
		return err
	}
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], k.Uint())
	_, err := w.Write(buf[8-width:])
	return err
}

func readMapKey(r io.Reader, dst reflect.Value, width uint8) (int64, error) {
	if width == 0 {
		var s String16
		n, err := (&s).ReadFrom(r)
		if err != nil {
			return n, err
		}
		dst.SetString(string(s))
		return n, nil
	}
	var buf [8]byte
	read, err := io.ReadFull(r, buf[8-width:])
	if err != nil {
		return int64(read), err
	}
	dst.SetUint(binary.BigEndian.Uint64(buf[:]))
	return int64(read), nil
}

// addressableMapValue copies a non-addressable MapIndex result into an
// addressable slot so pointer-receiver methods resolve through reflection.
// Go maps return non-addressable Values from MapIndex (rehashes may move
// slots), which hides the pointer method set from objectify.
func addressableMapValue(v reflect.Value) reflect.Value {
	addr := reflect.New(v.Type()).Elem()
	addr.Set(v)
	return addr
}

func (val mapValue) MarshalJSON() ([]byte, error) {
	if val.IsNil() {
		return jsonNull, nil
	}

	keyKind := val.Type().Key().Kind()
	if _, ok := supportedMapKey(keyKind); !ok {
		return nil, fmt.Errorf("map_value: unsupported key kind %s", keyKind)
	}

	var jmap = map[string]json.RawMessage{}

	for _, mapKey := range val.MapKeys() {
		var key string
		switch keyKind {
		case reflect.String:
			key = mapKey.String()
		default:
			key = strconv.FormatUint(mapKey.Uint(), 10)
		}

		o, err := objectify(addressableMapValue(val.MapIndex(mapKey)))
		if err != nil {
			return nil, err
		}

		value, err := o.MarshalJSON()
		if err != nil {
			return nil, err
		}

		jmap[key] = value
	}

	return json.Marshal(jmap)
}

func (val mapValue) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, jsonNull) {
		val.SetZero()
		return nil
	}

	keyType := val.Type().Key()
	if _, ok := supportedMapKey(keyType.Kind()); !ok {
		return fmt.Errorf("map_value: unsupported key kind %s", keyType.Kind())
	}

	var fields map[string]json.RawMessage
	err := json.Unmarshal(data, &fields)
	if err != nil {
		return err
	}

	// why: defer MakeMap until first entry so an empty JSON object leaves the map nil,
	// matching the binary path (mapValue.ReadFrom and runtime_map.go's slow path both
	// SetZero on count==0).
	if len(fields) == 0 {
		val.SetZero()
		return nil
	}
	val.Set(reflect.MakeMapWithSize(val.Type(), len(fields)))

	for k, v := range fields {
		mapKey := reflect.New(keyType).Elem()
		switch keyType.Kind() {
		case reflect.String:
			mapKey.SetString(k)
		default:
			// why: parse at the actual width so an overflowing literal errors instead
			// of silently truncating via reflect.SetUint.
			bitSize := int(keyType.Size()) * 8
			i, perr := strconv.ParseUint(k, 10, bitSize)
			if perr != nil {
				return fmt.Errorf("map_value: parse uint key %q: %w", k, perr)
			}
			mapKey.SetUint(i)
		}

		mapVal := reflect.New(val.Type().Elem()).Elem()
		o, oerr := objectify(mapVal)
		if oerr != nil {
			return oerr
		}

		if err = o.UnmarshalJSON(v); err != nil {
			return err
		}

		val.SetMapIndex(mapKey, mapVal)
	}

	return nil
}
