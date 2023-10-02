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

// Container provides access to the underlying data within the file.
type Container struct {
	// Name of the data container the name of the underlying file without file extension
	name string
	// File is the underlying file representation
	// which this container is based on.
	File File
	// Data is the decoded content of the file, represented as [Map].
	Data Map
	// A container can only have one root key which
	// must match the filename (without extension) of the FileProvider.
	// The RootKey is transparently added to all paths when working with the container.
	RootKey string
	// The Codec used to en-/decode the contents of the file into [Map].
	Codec FileCodec
}

// NewContainer creates a new container based on the given DataFileProvider.
// The referenceDirs are used to resolve YAML aliases and anchors
func NewContainer(file File, codec FileCodec) (*Container, error) {
	data, err := codec.Unmarshal(file.Bytes())
	if err != nil {
		return nil, err
	}

	// name of the container is the basename of the file
	name := filepath.Base(file.Path())
	name = name[:len(name)-len(filepath.Ext(name))]

	container := &Container{
		name:    name,
		File:    file,
		Data:    data,
		RootKey: file.BaseName(),
		Codec:   codec,
	}

	err = data.Walk(func(value interface{}, path Path) error {
		if !strings.EqualFold(path.First(), container.RootKey) {
			return fmt.Errorf("invalid root key: expected '%s', got '%s'", container.RootKey, path.First())
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return container, nil
}

// TODO: wildcard support (e.g. 'foo.*')
func (container *Container) Get(path Path) (val interface{}, err error) {
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

// TODO: if the first path segment is NOT the root key, make sure to append it (maybe this should be part of the container)
func (container *Container) Set(path Path, value interface{}) error {

	//
	byteValue, err := container.Codec.Marshal(value)
	if err != nil {
		return err
	}
	mapValue, err := container.Codec.Unmarshal(byteValue)
	if err != nil {
		return err
	}
	_ = mapValue

	return nil
}

// AllPaths traverses over all data of the container and returns all Paths to values.
func (container *Container) AllPaths() []Path {
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

func (container *Container) HasPath(path Path) bool {
	if _, err := container.Get(path); err != nil {
		return false
	}
	return true
}

func (container *Container) Name() string {
	return container.name
}
