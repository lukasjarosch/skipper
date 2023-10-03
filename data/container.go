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
	ErrEmptyRootKey            = fmt.Errorf("empty root key")
	ErrNilData                 = fmt.Errorf("data is nil")
	ErrNilCodec                = fmt.Errorf("codec is nil")
	ErrCannotSetNewPathTooDeep = fmt.Errorf("cannot set path with new path longer than one path segment")
)

type RawContainer struct {
	// Name of the data container the name of the underlying file without file extension
	// The name must also be the root key within the data [Map] in order to be consistent.
	name string
	// Data is the decoded content of the file, represented as [Map].
	Data Map
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

	container := &RawContainer{
		name:  name,
		Codec: codec,
		Data:  dataMap,
	}

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

func (container *RawContainer) Get(path Path) (val interface{}, err error) {
	// We support wildcard paths where the wildcard segment is the last path segment
	// For example: `foo.bar.*` will return anything under `foo.bar`
	// Currently inline wildcards (e.g. `foo.*.baz`) are not supported.
	if path.Last() == WildcardIdentifier {
		newPath := NewPathVar(path[:len(path)-1]...)
		val, err := container.Data.GetPath(newPath)
		if err != nil {
			return nil, ErrPathNotFound{Path: path, Err: err}
		}
		return val, nil
	}

	val, err = container.Data.GetPath(path)
	if err != nil {
		return nil, ErrPathNotFound{Path: path, Err: err}
	}

	return val, nil
}

// AllPaths traverses over all data of the container and returns all Paths to values.
func (container *RawContainer) AllPaths() []Path {
	// we're using a map to avoid the pescy append slice behaviour
	pathMap := make(map[string]bool)
	container.Data.Walk(func(value interface{}, path Path) error {
		pathMap[path.String()] = true
		return nil
	})

	var paths []Path
	for p := range pathMap {
		paths = append(paths, NewPath(p))
	}

	return paths
}

func (container *RawContainer) HasPath(path Path) bool {
	if _, err := container.Get(path); err != nil {
		return false
	}
	return true
}

// TODO: if the first path segment is NOT the root key, make sure to append it (maybe this should be part of the container)
func (container *RawContainer) Set(path Path, value interface{}) error {
	if !container.Data.CanSetPath(path) {
		return fmt.Errorf("%w: %s", ErrCannotSetNewPathTooDeep, path)
	}

	// attempt to marshal and unmarshal the value in case
	// it is a complex type which needs to be encoded properly
	// this function will not return an error but only the unmodified input value
	encode := func(in interface{}) interface{} {
		byteValue, err := container.Codec.Marshal(value)
		if err != nil {
			return in
		}
		mapValue, err := container.Codec.Unmarshal(byteValue)
		if err != nil {
			return in
		}
		return mapValue
	}

	err := container.Data.SetPath(path, encode(value))
	if err != nil {
		return err
	}
	return nil
}

func (container *RawContainer) Name() string {
	return container.name
}

// FileContainer provides access to the underlying data within the file.
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
