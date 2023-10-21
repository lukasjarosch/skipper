package data

import (
	"fmt"
	"reflect"
	"strconv"
)

var (
	ErrExpectedNumericArrayIndex = fmt.Errorf("expected numeric array index")
	ErrKeyNotFound               = fmt.Errorf("key not found")
	ErrInvalidValue              = fmt.Errorf("invalid value")
	ErrUnsupportedDataType       = fmt.Errorf("unsupported data type")
	ErrNegativeIndex             = fmt.Errorf("negative index")
	ErrTypeChange                = fmt.Errorf("changing types is not supported")
)

type WalkFunc func(path Path, data interface{}, isLeaf bool) error

func Walk(data interface{}, walkFn WalkFunc) error {
	return walk(data, Path{}, walkFn)
}

// traverse implements a basic DFS and traverses over every
// node in the map, calling the TraverseFunc on each of them.
func walk(parent interface{}, path Path, walkFn WalkFunc) error {
	if parent == nil {
		return nil
	}

	var parentValue = reflect.ValueOf(parent)

	switch parentValue.Kind() {
	case reflect.Map:
		// ignore the root (empty) path and only recurse into 'actual paths'
		if len(path) > 0 {
			if err := walkFn(path, parent, false); err != nil {
				return err
			}
		}
		for _, key := range parentValue.MapKeys() {
			value := parentValue.MapIndex(key)
			subPath := path.Append(fmt.Sprintf("%v", key.Interface()))

			if err := walk(value.Interface(), subPath, walkFn); err != nil {
				return err
			}
		}
	case reflect.Array, reflect.Slice:
		// ignore the root (empty) path and only recurse into 'actual paths'
		if len(path) > 0 {
			if err := walkFn(path, parent, false); err != nil {
				return err
			}
		}

		for i := 0; i < parentValue.Len(); i++ {
			value := parentValue.Index(i)
			subPath := path.Append(fmt.Sprintf("%v", i))

			if err := walk(value.Interface(), subPath, walkFn); err != nil {
				return err
			}
		}
	default:
		return walkFn(path, parent, true)
	}

	return nil
}

func Get(data interface{}, key string) (interface{}, error) {
	data = ResolveValue(data)

	if IsZero(data) {
		return nil, ErrNilData
	}

	if IsMap(data) {
		if valueV := reflect.ValueOf(data).MapIndex(reflect.ValueOf(key)); valueV.IsValid() {
			if valueI := valueV.Interface(); !IsZero(valueI) {
				return valueI, nil
			}
		}
	}

	if IsArray(data) {
		if !IsInteger(key) {
			return nil, ErrExpectedNumericArrayIndex
		}
		index, _ := strconv.Atoi(key)

		if index >= 0 && index < reflect.ValueOf(data).Len() {
			return reflect.ValueOf(data).Index(index).Interface(), nil
		}
	}

	return nil, ErrKeyNotFound
}

func DeepGet(data interface{}, path Path) (interface{}, error) {
	current := ResolveValue(data)

	for i := 0; i < len(path); i++ {
		var pathSegment = path[i]
		var curValue = reflect.ValueOf(current)

		if !curValue.IsValid() {
			return nil, ErrInvalidValue
		}

		var curTyp = curValue.Type()

		// for pointers and interfaces, get the underlying type
		switch curValue.Kind() {
		case reflect.Interface, reflect.Ptr:
			curTyp = curTyp.Elem()
		}

		switch curTyp.Kind() {
		case reflect.Array, reflect.Slice:
			if !IsInteger(pathSegment) {
				return nil, ErrExpectedNumericArrayIndex
			}
			segmentIndex, err := strconv.Atoi(pathSegment)
			if err != nil {
				return nil, fmt.Errorf("%w: %w", ErrExpectedNumericArrayIndex, err)
			}
			if segmentIndex >= curValue.Len() {
				return nil, fmt.Errorf("index '%d' is too large for array of length '%d'", segmentIndex, curValue.Len())
			}

			value := curValue.Index(segmentIndex).Interface()
			if value == nil {
				return nil, fmt.Errorf("unexpected nil value at pathSegment '%d'", segmentIndex)
			}

			current = value
			continue

		case reflect.Map:
			mapVal := curValue.MapIndex(reflect.ValueOf(pathSegment))
			if !mapVal.IsValid() {
				return nil, fmt.Errorf("unexpected invalid value at pathSegment '%s'", pathSegment)
			}
			current = mapVal.Interface()
			continue
		default:
			return nil, fmt.Errorf("attempted to retrieve nested value of scalar")
		}
	}

	return current, nil
}

func Set(data interface{}, key string, value interface{}) (interface{}, error) {
	if data == nil {
		return nil, ErrNilData
	}
	if key == "" {
		return nil, fmt.Errorf("key cannot be empty")
	}

	switch v := data.(type) {
	case map[string]interface{}:
		v[key] = value
		return v, nil
	case []interface{}:
		idx, err := strconv.Atoi(key)
		if err != nil {
			return nil, fmt.Errorf("non-integer index for slice")
		}

		// ensure the slice is large enough by extending it if needed.
		for len(v) <= idx {
			v = append(v, nil)
		}

		v[idx] = value
		return v, nil

	default:
		return nil, ErrUnsupportedDataType
	}
}

func DeepSet(data interface{}, path Path, value interface{}) (interface{}, error) {
	if len(path) == 0 {
		return nil, ErrEmptyPath
	}

	currentSegment := path[0]
	remainingSegments := path[1:]

	switch currentData := data.(type) {
	case []interface{}:
		index, err := strconv.Atoi(currentSegment)
		if err != nil {
			return nil, fmt.Errorf("cannot use non-integer key in slice: %w", ErrTypeChange)
		}

		if index < 0 {
			return nil, ErrNegativeIndex
		}

		for len(currentData) <= index {
			currentData = append(currentData, nil)
		}

		if len(remainingSegments) == 0 {
			currentData[index] = value
			return currentData, nil
		}

		updated, err := DeepSet(currentData[index], remainingSegments, value)
		if err != nil {
			return nil, err
		}

		currentData[index] = updated
		return currentData, nil

	case map[string]interface{}:
		if len(remainingSegments) == 0 {
			dataValue, err := Set(currentData, currentSegment, value)
			if err != nil {
				return nil, err
			}
			return dataValue, nil
		}

		updated, err := DeepSet(currentData[currentSegment], remainingSegments, value)
		if err != nil {
			return nil, err
		}

		currentData[currentSegment] = updated
		return data, nil

	case nil:
		if IsInteger(currentSegment) {
			index, err := strconv.Atoi(currentSegment)
			if err != nil {
				return nil, fmt.Errorf("cannot use non-integer key in slice: %w", ErrTypeChange)
			}

			if index < 0 {
				return nil, ErrNegativeIndex
			}

			// at this point we know that the - currently 'nil' - value should actually be a []interface{}
			// which has at least the length 'index+1'; let's create that
			newValue := make([]interface{}, index+1)

			// if there are no more path segments, set the target value and return
			if len(remainingSegments) == 0 {
				newValue[index] = value
				return newValue, nil
			}

			// otherwise, recurse deeper...
			updated, err := DeepSet(newValue[index], remainingSegments, value)
			if err != nil {
				return nil, err
			}

			newValue[index] = updated
			return newValue, nil
		}

		newValue := map[string]interface{}{}

		// if there are no more path segments, set the target value and return
		if len(remainingSegments) == 0 {
			newValue[currentSegment] = value
			return newValue, nil
		}

		// there are more path segments
		updated, err := DeepSet(newValue[currentSegment], remainingSegments, value)
		if err != nil {
			return nil, err
		}

		newValue[currentSegment] = updated
		return newValue, nil

	default:
		// At this point 'value' can be either something we don't support like 'struct'
		// or any kind of scalar value (e.g. 5)

		// First we'll eliminate anything we don't support
		if IsKind(ResolveValue(data), UnsupportedTypes...) {
			return nil, ErrUnsupportedDataType
		}

		// Any remaining types should be scalar values.
		// We could go ahead and just re-type the value to the required target
		// value by evaluating the currentSegment just as above.
		// But allowing to change scalar types would also mean that
		// we should allow changing []interface{} to map[string]interface{}.
		// Hence we simply don't allow chaning types at the moment.
		return nil, ErrTypeChange
	}
}

func Merge(baseData map[string]interface{}, mergeData map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(baseData))
	for k, v := range baseData {
		result[k] = v
	}

	for k, v := range mergeData {
		if v, ok := v.(map[string]interface{}); ok {
			if bv, ok := result[k]; ok {
				if bv, ok := bv.(map[string]interface{}); ok {
					result[k] = Merge(bv, v)
					continue
				}
			}
		}
		if v, ok := v.([]interface{}); ok {
			if bv, ok := result[k]; ok {
				if bv, ok := bv.([]interface{}); ok {
					result[k] = append(bv, v...)
					continue
				}
			}
		}
		result[k] = v
	}

	return result
}

type PathSelectorFunc func(data interface{}, path Path, isLeaf bool) bool

var SelectAllPaths PathSelectorFunc = func(_ interface{}, path Path, isLeaf bool) bool {
	return true
}
var SelectLeafPaths PathSelectorFunc = func(_ interface{}, path Path, isLeaf bool) bool {
	if isLeaf {
		return true
	}
	return false
}

func Paths(data map[string]interface{}, selectorFn PathSelectorFunc) []Path {
	pathSelectMap := make(map[string]bool)

	err := Walk(data, func(path Path, _ interface{}, isLeaf bool) error {
		pathSelectMap[path.String()] = selectorFn(data, path, isLeaf)
		return nil
	})
	if err != nil {
		return nil
	}

	paths := []Path{}
	for p, selected := range pathSelectMap {
		if !selected {
			continue
		}
		paths = append(paths, NewPath(p))
	}

	return paths
}
