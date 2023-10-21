package data

import (
	"fmt"
	"reflect"
	"strconv"
	"unicode/utf8"
)

var SliceTypes = []reflect.Kind{
	reflect.Slice,
	reflect.Array,
}

var UnsupportedTypes = []reflect.Kind{
	reflect.Invalid,
	reflect.Complex64,
	reflect.Complex128,
	reflect.Chan,
	reflect.Func,
	reflect.Interface,
	reflect.Ptr,
	reflect.Struct,
}

var CompoundTypes = []reflect.Kind{
	reflect.Invalid,
	reflect.Complex64,
	reflect.Complex128,
	reflect.Array,
	reflect.Chan,
	reflect.Func,
	reflect.Interface,
	reflect.Map,
	reflect.Ptr,
	reflect.Slice,
	reflect.Struct,
}

func ToString(in interface{}) (string, error) {
	if in == nil {
		return ``, nil
	} else if err, ok := in.(error); ok {
		return err.Error(), nil
	} else if s, ok := in.(fmt.Stringer); ok {
		return s.String(), nil
	}

	var asBytes []byte

	if u8, ok := in.([]uint8); ok {
		asBytes = []byte(u8)
	} else if b, ok := in.([]byte); ok {
		asBytes = b
	} else if r, ok := in.([]rune); ok {
		return string(r), nil
	}

	if len(asBytes) > 0 {
		if out := string(asBytes); utf8.ValidString(out) {
			return out, nil
		} else {
			return ``, fmt.Errorf("Given %T is not a valid UTF-8 string", in)
		}
	}

	if inT := reflect.TypeOf(in); inT != nil {
		switch inT.Kind() {
		case reflect.Float32:
			return strconv.FormatFloat(reflect.ValueOf(in).Float(), 'f', -1, 32), nil
		case reflect.Float64:
			return strconv.FormatFloat(reflect.ValueOf(in).Float(), 'f', -1, 64), nil
		case reflect.Bool:
			return strconv.FormatBool(in.(bool)), nil
		case reflect.String:
			if inStr, ok := in.(string); ok {
				return inStr, nil
			}
		}

		if !IsKind(in, CompoundTypes...) {
			return fmt.Sprintf("%v", in), nil
		}
	}

	return ``, fmt.Errorf("Unable to convert type '%T' to string", in)
}

func IsInteger(in interface{}) bool {
	inV := reflect.ValueOf(in)

	switch inV.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return true

	default:
		if asStr, err := ToString(in); err == nil {
			if _, err := strconv.Atoi(asStr); err == nil {
				return true
			}
		}
	}

	return false
}

// Dectect whether the concrete underlying value of the given input is one or more
// Kinds of value.
func IsKind(in interface{}, kinds ...reflect.Kind) bool {
	var inT reflect.Type

	if v, ok := in.(reflect.Value); ok && v.IsValid() {
		inT = v.Type()
	} else if v, ok := in.(reflect.Type); ok {
		inT = v
	} else {
		in = ResolveValue(in)
		inT = reflect.TypeOf(in)
	}

	if inT == nil {
		return false
	}

	for _, k := range kinds {
		if inT.Kind() == k {
			return true
		}
	}

	return false
}

func ResolveValue(in interface{}) interface{} {
	var inV reflect.Value

	if vV, ok := in.(reflect.Value); ok {
		inV = vV
	} else {
		inV = reflect.ValueOf(in)
	}

	if inV.IsValid() {
		if inT := inV.Type(); inT == nil {
			return nil
		}

		switch inV.Kind() {
		case reflect.Ptr, reflect.Interface:
			return ResolveValue(inV.Elem())
		}

		in = inV.Interface()
	}

	return in
}

// Returns whether the given value represents the underlying type's zero value
func IsZero(value interface{}) bool {
	if value == nil {
		return true
	} else if valueV, ok := value.(reflect.Value); ok && valueV.IsValid() {
		if valueV.CanInterface() {
			value = valueV.Interface()
		}
	}

	return reflect.DeepEqual(
		value,
		reflect.Zero(reflect.TypeOf(value)).Interface(),
	)
}

// Returns whether the given value is a slice or array.
func IsArray(in interface{}) bool {
	return IsKind(in, SliceTypes...)
}

// Returns whether the given value is a map.
func IsMap(in interface{}) bool {
	return IsKind(in, reflect.Map)
}
