package expression

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"

	"github.com/iancoleman/strcase"
)

type FuncMap map[string]any

var errorType = reflect.TypeFor[error]()

var builtins = FuncMap{
	"get_env": get_env,
	"set_env": set_env,
	"default": defaultFunc,

	// string casing helpers
	"to_upper":               func(s string) string { return strings.ToUpper(s) },
	"to_lower":               func(s string) string { return strings.ToLower(s) },
	"to_snake":               strcase.ToSnake,
	"to_screaming_snake":     strcase.ToScreamingSnake,
	"to_camel":               strcase.ToCamel,
	"to_kebab":               strcase.ToKebab,
	"to_screaming_kebab":     strcase.ToScreamingKebab,
	"to_lower_camel":         strcase.ToLowerCamel,
	"to_delimited":           to_delimited,
	"to_screaming_delimited": to_screaming_delimited,
}

var builtinFuncsOnce struct {
	sync.Once
	v map[string]reflect.Value
}

func builtinFuncs() map[string]reflect.Value {
	builtinFuncsOnce.Do(func() {
		builtinFuncsOnce.v = createValueFuncs(builtins)
	})
	return builtinFuncsOnce.v
}

// createValueFuncs turns a FuncMap into a map[string]reflect.Value
func createValueFuncs(funcMap FuncMap) map[string]reflect.Value {
	m := make(map[string]reflect.Value)
	addValueFuncs(m, funcMap)
	return m
}

// addValueFuncs adds to values the functions in funcs, converting them to reflect.Values.
func addValueFuncs(out map[string]reflect.Value, in FuncMap) {
	for name, fn := range in {
		v := reflect.ValueOf(fn)
		if v.Kind() != reflect.Func {
			panic("value for " + name + " not a function")
		}
		if !goodFunc(v.Type()) {
			panic(fmt.Errorf("can't install method/function %q with %d results", name, v.Type().NumOut()))
		}
		out[name] = v
	}
}

// goodFunc returns true when the given function has either one return value
// or two, whereas the second must be of type 'error'.
// All other function signatures are not good and false is returned.
func goodFunc(fn reflect.Type) bool {
	switch {
	case fn.NumOut() == 1:
		return true
	case fn.NumOut() == 2 && fn.Out(1) == errorType:
		return true
	}

	return false
}

func findFunction(name string, s *state) (reflect.Value, bool) {
	if fn := s.funcMap[name]; reflect.ValueOf(fn).IsValid() {
		return reflect.ValueOf(fn), true
	}
	if fn := builtinFuncs()[name]; fn.IsValid() {
		return fn, true
	}

	return zero, false
}

func safeCall(ident string, fun reflect.Value, args []reflect.Value) (val reflect.Value, err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("recovered panic in %s: %v", ident, r)
			}
		}
	}()

	ret := fun.Call(args)
	if len(ret) == 2 && !ret[1].IsNil() {
		return ret[0], fmt.Errorf("%s: %w", ident, ret[1].Interface().(error))
	}
	return ret[0], nil
}

// get_env will lookup the given name as environment variable and return its value.
// If the variable does not exist, an error is returned.
// If the variable exists, but is empty, the empty value is returned.
func get_env(name string) (string, error) {
	val, exists := os.LookupEnv(name)
	if !exists {
		return "", fmt.Errorf("environment variable not set: %s", name)
	}
	return val, nil
}

// set_env will attempt to set an environment variable with the given name and value.
// It will return the set value and an error (if any).
func set_env(name string, value string) (string, error) {
	return value, os.Setenv(name, value)
}

// defaultFunc checks whether 'input' is set and returns defaultStr if not set.
// If 'input' is set, 'input' is returned.
func defaultFunc(input, dfault interface{}) interface{} {
	empty := func(given interface{}) bool {
		g := reflect.ValueOf(given)
		if !g.IsValid() {
			return true
		}

		// Basically adapted from text/template.isTrue
		switch g.Kind() {
		default:
			return g.IsNil()
		case reflect.Array, reflect.Slice, reflect.Map, reflect.String:
			return g.Len() == 0
		case reflect.Bool:
			return !g.Bool()
		case reflect.Complex64, reflect.Complex128:
			return g.Complex() == 0
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return g.Int() == 0
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			return g.Uint() == 0
		case reflect.Float32, reflect.Float64:
			return g.Float() == 0
		case reflect.Struct:
			return false
		}
	}

	if empty(input) {
		return dfault
	}

	return input
}

// to_delimited is a wrapper for [strcase.ToDelimited].
// If the delim is empty, '.' is used.
func to_delimited(str, delim string) string {
	if delim == "" {
		delim = "."
	}

	return strcase.ToDelimited(str, byte(delim[0]))
}

// to_delimited is a wrapper for [strcase.ToScreamingDelimited].
// If the delim is empty, '.' is used.
func to_screaming_delimited(str, delim string) string {
	if delim == "" {
		delim = "."
	}

	return strcase.ToScreamingDelimited(str, byte(delim[0]), "", true)
}
