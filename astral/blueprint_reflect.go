package astral

import (
	"fmt"
	"reflect"
)

var objectInterface = reflect.TypeOf((*Object)(nil)).Elem()

// BlueprintFromType derives a Blueprint from a struct reflect.Type by inspecting its exported
// fields. The struct (or its pointer) must implement Object so its ObjectType() can be used
// as the blueprint's Type.
//
// If the prototype satisfies PrimitiveAlias, an alias-kind Blueprint is returned: Underlying is
// set to UnderlyingPrimitive() and Fields stays empty. Otherwise a struct-kind Blueprint
// is returned, with each exported field mapped to a Spec carrier:
//
//	implements Object & primitive allowlist → *PrimitiveSpec
//	implements Object & not primitive       → *RefSpec
//	*T                                      → *PtrSpec{Type: T.ObjectType()}
//	[]T                                     → *SliceSpec{Type: elem name or "" for Object}
//	map[K]V                                 → *MapSpec{KeyType: key name, ValueType: value name or ""}
//	Object (interface)                      → *ObjectSpec
func BlueprintFromType(t reflect.Type) (*Blueprint, error) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	typeName, typeNameErr := concreteObjectTypeOf(t)

	// why: PrimitiveAlias is the Go-side declaration of an alias-kind Blueprint and the
	// underlying Go type may be a primitive newtype (e.g. `type Mode astral.Uint8`),
	// not a struct. Probe before the struct check so non-struct PrimitiveAlias prototypes
	// derive correctly.
	if a, ok := reflect.New(t).Interface().(PrimitiveAlias); ok {
		if typeNameErr != nil {
			return nil, fmt.Errorf("BlueprintFromType: %w", typeNameErr)
		}
		bp := &Blueprint{Type: String16(typeName), Underlying: String16(a.UnderlyingPrimitive())}
		if err := validateBlueprint(bp); err != nil {
			return nil, fmt.Errorf("BlueprintFromType: %w", err)
		}
		return bp, nil
	}

	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("BlueprintFromType: want struct or *struct, got %s", t)
	}

	if typeNameErr != nil {
		return nil, fmt.Errorf("BlueprintFromType: %w", typeNameErr)
	}

	bp := &Blueprint{Type: String16(typeName)}

	emptyObject := reflect.TypeOf(EmptyObject{})
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if !sf.IsExported() {
			continue
		}
		// why: EmptyObject is the framework's no-payload marker (embedded in Ack/EOS/Nil
		// and similar). It carries no wire bytes, so it has nothing to describe in a Spec.
		if sf.Type == emptyObject {
			continue
		}
		spec, err := specFromType(sf.Type)
		if err != nil {
			return nil, fmt.Errorf("BlueprintFromType %s.%s: %w", typeName, sf.Name, err)
		}
		bp.Fields = append(bp.Fields, Field{
			Name: String16(sf.Name),
			Spec: spec,
		})
	}

	if err := validateBlueprint(bp); err != nil {
		return nil, fmt.Errorf("BlueprintFromType: %w", err)
	}

	return bp, nil
}

// BlueprintOf is a convenience wrapper around BlueprintFromType that uses the runtime type of v.
func BlueprintOf(v Object) (*Blueprint, error) {
	return BlueprintFromType(reflect.TypeOf(v))
}

// specFromType maps a Go field type to its corresponding Spec carrier.
//
// Dispatch order matters: pointers and the Object interface are container-shaped regardless
// of whether they implement Object themselves. Anything else is first probed for an Object
// implementation — astral types built on a slice or map (Bytes32 = []byte, etc.) are leaf
// primitives, not generic collections, and must short-circuit before the Slice/Map dispatch.
func specFromType(t reflect.Type) (Spec, error) {
	if t.Kind() == reflect.Interface {
		// why: accepting any interface whose method set embeds astral.Object lets
		// sub-interfaces like exonet.Endpoint flow through an ObjectSpec slot, since
		// every concrete value carries its astral type tag on the wire.
		if t.Implements(objectInterface) {
			return &ObjectSpec{}, nil
		}
		return nil, fmt.Errorf("interface %s does not embed astral.Object", t)
	}

	if t.Kind() == reflect.Ptr {
		name, err := concreteObjectTypeOf(t.Elem())
		if err != nil {
			return nil, fmt.Errorf("pointer elem: %w", err)
		}
		return &PtrSpec{Type: String16(name)}, nil
	}

	// Probe Object implementation before falling through to Slice/Map dispatch so that
	// astral leaf types like Bytes32 (type Bytes32 []byte) are recognized as primitives
	// rather than generic []byte slices.
	if name, ok := tryObjectType(t); ok {
		if isAllowedPrimitive(name) {
			return &PrimitiveSpec{PrimitiveType: String16(name)}, nil
		}
		return &RefSpec{Type: String16(name)}, nil
	}

	switch t.Kind() {
	case reflect.Slice:
		elemName, err := elemTypeName(t.Elem())
		if err != nil {
			return nil, fmt.Errorf("slice elem: %w", err)
		}
		return &SliceSpec{Type: String16(elemName)}, nil

	case reflect.Array:
		elemName, err := elemTypeName(t.Elem())
		if err != nil {
			return nil, fmt.Errorf("array elem: %w", err)
		}
		return &ArraySpec{Type: String16(elemName), Length: Uint32(t.Len())}, nil

	case reflect.Map:
		keyName, err := mapKeyTypeName(t.Key())
		if err != nil {
			return nil, err
		}
		valueName, err := elemTypeName(t.Elem())
		if err != nil {
			return nil, fmt.Errorf("map value: %w", err)
		}
		return &MapSpec{KeyType: String16(keyName), ValueType: String16(valueName)}, nil
	}

	return nil, fmt.Errorf("type %s does not implement Object and is not a supported container", t)
}

// tryObjectType returns the ObjectType() of a non-interface type or ("", false) if the type
// does not implement Object. Unlike concreteObjectTypeOf it does not error — it's used to
// probe whether a Kind should be treated as a leaf primitive/ref or fall through to the
// container dispatch.
func tryObjectType(t reflect.Type) (string, bool) {
	if t.Kind() == reflect.Interface {
		return "", false
	}
	if t.Kind() == reflect.Ptr {
		if o, ok := reflect.New(t.Elem()).Interface().(Object); ok {
			if name := o.ObjectType(); name != "" {
				return name, true
			}
		}
		return "", false
	}
	if o, ok := reflect.New(t).Elem().Interface().(Object); ok {
		if name := o.ObjectType(); name != "" {
			return name, true
		}
	}
	if o, ok := reflect.New(t).Interface().(Object); ok {
		if name := o.ObjectType(); name != "" {
			return name, true
		}
	}
	return "", false
}

// concreteObjectTypeOf returns the ObjectType of a non-interface Go type by constructing a
// zero instance. Tries value-receiver Object methods first, then falls back to pointer.
func concreteObjectTypeOf(t reflect.Type) (string, error) {
	if t.Kind() == reflect.Interface {
		return "", fmt.Errorf("expected concrete type, got interface %s", t)
	}

	// pointer types: instantiate *Elem directly
	if t.Kind() == reflect.Ptr {
		if o, ok := reflect.New(t.Elem()).Interface().(Object); ok {
			name := o.ObjectType()
			if name == "" {
				return "", fmt.Errorf("type %s implements Object but ObjectType() returned empty", t)
			}
			return name, nil
		}
		return "", fmt.Errorf("type %s does not implement Object", t)
	}

	// value types: try value-receiver, then pointer-receiver
	if o, ok := reflect.New(t).Elem().Interface().(Object); ok {
		name := o.ObjectType()
		if name == "" {
			return "", fmt.Errorf("type %s implements Object but ObjectType() returned empty", t)
		}
		return name, nil
	}
	if o, ok := reflect.New(t).Interface().(Object); ok {
		name := o.ObjectType()
		if name == "" {
			return "", fmt.Errorf("type %s implements Object but ObjectType() returned empty", t)
		}
		return name, nil
	}
	return "", fmt.Errorf("type %s does not implement Object", t)
}

// elemTypeName returns "" when the element is the Object interface (heterogeneous container),
// otherwise the concrete element's ObjectType.
func elemTypeName(t reflect.Type) (string, error) {
	if t.Kind() == reflect.Interface {
		if t.Implements(objectInterface) {
			return "", nil
		}
		return "", fmt.Errorf("interface %s does not embed astral.Object", t)
	}
	return concreteObjectTypeOf(t)
}

// mapKeyTypeName translates a Go map key reflect.Type to the canonical wire name accepted by
// MapSpec.KeyType (must be in mapKeyAllowlist).
func mapKeyTypeName(t reflect.Type) (string, error) {
	switch t.Kind() {
	case reflect.String:
		return "string16", nil
	case reflect.Uint8:
		return "uint8", nil
	case reflect.Uint16:
		return "uint16", nil
	case reflect.Uint32:
		return "uint32", nil
	case reflect.Uint64:
		return "uint64", nil
	}
	// why: reflect.Uint is rejected here to stay aligned with supportedMapKey in map_value.go —
	// platform-dependent width would split content hashes across architectures. See the same
	// rejection in objectify for non-map fields.
	return "", fmt.Errorf("unsupported map key type %s (must be fixed-width unsigned int or string)", t)
}
