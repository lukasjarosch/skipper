package data

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

// File is an interface that represents a file and provides methods
// to access its information required by the [Container].
type File interface {
	// Bytes returns the content of the file as a byte slice.
	Bytes() []byte
	// Mode returns the file mode of the underlying file.
	Mode() fs.FileMode
	// Path returns the path of the file.
	Path() string
	// BaseName returns the base name of the file (file name without extension).
	BaseName() string
}

// FileCodec is used to de-/encode the content of a [File].
type FileCodec interface {
	// DecodeBytes decodes a slice of bytes and returns a generic [Map]
	// which we can then use to work on the data.
	Unmarshal([]byte) (Map, error)
	// UnmarshalTarget decodes a slice of bytes into the given target interface
	UnmarshalTarget([]byte, interface{}) error
	// UnmarshalPath uses [FileCodec.Unmarshal], [FileCodec.Marshal] [FileCodec.UnmarshalPath]
	// and [Map.GetPath] to decode the given data into a [Map], resolve the given path
	// into it, and then unmarshal the resulting value into the target interface.
	UnmarshalPath([]byte, Path, interface{}) error
	// Marshal encodes the given interface into a slice of bytes
	Marshal(in interface{}) ([]byte, error)
}

var (
	ErrEmptyContainerName      = fmt.Errorf("container name empty")
	ErrCannotSetRootKey        = fmt.Errorf("cannot set the root key of a container")
	ErrEmptyRootKey            = fmt.Errorf("empty root key")
	ErrNilData                 = fmt.Errorf("data is nil")
	ErrNilCodec                = fmt.Errorf("codec is nil")
	ErrCannotSetNewPathTooDeep = fmt.Errorf("cannot set path with new path longer than one path segment")
	ErrInlineWildcard          = fmt.Errorf("inline wildcard paths are not supported")
)

type RawContainer struct {
	// Name of the data container the name of the underlying file without file extension
	// The name must also be the root key within the data [Map] in order to be consistent.
	name string
	// data is the decoded content of the file, represented as [Map].
	data Map
	// The Codec used to en-/decode the contents of the file into [Map].
	Codec FileCodec
}

func NewRawContainer(name string, data interface{}, codec FileCodec) (*RawContainer, error) {
	if name == "" {
		return nil, ErrEmptyContainerName
	}
	if data == nil {
		return nil, ErrNilData
	}
	if codec == nil {
		return nil, ErrNilCodec
	}

	dataByte, err := codec.Marshal(data)
	if err != nil {
		return nil, err
	}

	dataMap, err := codec.Unmarshal(dataByte)
	if err != nil {
		return nil, err
	}

	container := &RawContainer{
		name:  name,
		Codec: codec,
		data:  dataMap,
	}

	// the only allowed root key of the underlying [Map] must be the same as the container name
	err = dataMap.Walk(func(value interface{}, path Path) error {
		if !strings.EqualFold(path.First(), container.name) {
			return fmt.Errorf("invalid root key: expected '%s', got '%s'", container.name, path.First())
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (container *RawContainer) Get(path Path) (val Value, err error) {
	if len(path) == 0 {
		return Value{}, ErrEmptyPath
	}

	// If the path consists only of a WildcardIdentifier, return the whole data map
	if len(path) == 1 && path.First() == WildcardIdentifier {
		return NewValue(container.data), nil
	}

	// For all other cases, make sure the containerName is the first path segment
	// in order to properly index into the data.
	// This also means that one can omit the containerName in the path
	if path.First() != container.name {
		path = path.Prepend(container.name)
	}

	// Wildcard paths in which the wildcard segment is the last path segment are also supported
	// For example: `foo.bar.*` will return anything under `foo.bar`
	if path.Last() == WildcardIdentifier {
		newPath := path.StripSuffix(path.LastSegment())
		val, err := container.data.Get(newPath)
		if err != nil {
			return Value{}, ErrPathNotFound{Path: path, Err: err}
		}
		return NewValue(val), nil
	}

	// Inline wildcards (e.g. `foo.*.baz`) are not supported.
	// Check if a [WildcardIdentifier] appears anywhere inside the Path except the last segment.
	if path.SegmentIndex(WildcardIdentifier) > 0 &&
		path.SegmentIndex(WildcardIdentifier) != len(path)-1 {
		return Value{}, ErrInlineWildcard
	}

	raw, err := container.data.Get(path)
	if err != nil {
		return Value{}, ErrPathNotFound{Path: path, Err: err}
	}

	return NewValue(raw), nil
}

func (container *RawContainer) MustGet(path Path) Value {
	val, err := container.Get(path)
	if err != nil {
		panic(err)
	}
	return val
}

func (container *RawContainer) HasPath(path Path) bool {
	if _, err := container.Get(path); err != nil {
		return false
	}
	return true
}

func (container *RawContainer) ValuePaths() []Path {
	return container.data.ValuePaths()
}

// attemptEncode will attempt to marshal and subsequently unmarshal the given interface
// into a [Map].
// If any of those operations fail, the input value is returned.
func (container *RawContainer) attemptEncode(in interface{}) interface{} {
	if in == nil {
		return nil
	}
	byteSlice, err := container.Codec.Marshal(in)
	if err != nil {
		return in
	}

	// if we get an empty byte slice 'Unmarshall' will
	// either return nothing usable to just nothin at all, abort here
	if len(byteSlice) == 0 {
		return in
	}

	mapValue, err := container.Codec.Unmarshal(byteSlice)
	if err != nil {
		return in
	}
	return mapValue
}

func (container *RawContainer) set(path Path, value Value, attemptEncode bool) error {
	if len(path) == 0 {
		return ErrEmptyPath
	}

	// we do not allow setting the whole container data
	// as it may change the rootKey which must match the containerName
	if len(path) == 1 && path.First() == WildcardIdentifier {
		return ErrCannotSetRootKey
	}

	if path.First() != container.name {
		path = path.Prepend(container.name)
	}

	if attemptEncode {
		value.Raw = container.attemptEncode(value.Raw)
	}

	err := container.data.Set(path, value.Raw)
	if err != nil {
		return err
	}
	return nil
}

func (container *RawContainer) Set(path Path, value Value) error {
	return container.set(path, value, true)
}

// SetRaw works just like [RawContainer.Set] with the difference that it accepts
// a raw interface to set at the specified path.
// The function will also NOT attempt to encode the value with the configured codec.
func (container *RawContainer) SetRaw(path Path, value Value) error {
	return container.set(path, value, false)
}

// TODO: this has a lot of edge cases and needs heavy testing
func (container *RawContainer) Merge(path Path, data Map) error {
	val, err := container.Get(path)
	if err != nil {
		return err
	}

	valMap, err := val.Map()
	if err != nil {
		return err
	}

	replaced := valMap.MergeReplace(data)

	// In case the path is a wildcard path, remove the identifier before
	// setting the path again because [RawContainer.SetRaw] does not support it.
	// We're only handling paths with length > 1. If the path
	// only consists of a WildcardIdentifier, the [Container.set] function wil handle it.
	if len(path) > 1 && path.Last() == WildcardIdentifier {
		path = path.StripSuffix(NewPath(WildcardIdentifier))
	}

	err = container.SetRaw(path, NewValue(replaced))
	if err != nil {
		return err
	}

	return nil
}

func (container *RawContainer) Name() string {
	return container.name
}

// FileContainer is a [RawContainer] which is based on a [File].
type FileContainer struct {
	*RawContainer
	// File is the underlying file representation
	// which this container is based on.
	File File
}

// NewFileContainer creates a new container based on the given DataFileProvider.
// The referenceDirs are used to resolve YAML aliases and anchors
func NewFileContainer(file File, codec FileCodec) (*FileContainer, error) {
	data, err := codec.Unmarshal(file.Bytes())
	if err != nil {
		return nil, err
	}

	// name of the container is the basename of the file
	name := filepath.Base(file.Path())
	name = name[:len(name)-len(filepath.Ext(name))]

	rawContainer, err := NewRawContainer(name, data, codec)
	if err != nil {
		return nil, err
	}

	container := &FileContainer{
		RawContainer: rawContainer,
		File:         file,
	}

	return container, nil
}
