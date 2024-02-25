package skipper

import (
	"errors"
	"fmt"

	"github.com/lukasjarosch/skipper/data"
)

var (
	ErrClassIdentifierDoesNotExist = fmt.Errorf("class identifier does not exist")
	ErrClassAlreadyRegistered      = fmt.Errorf("class already registered")
	ErrClassDoesNotExist           = fmt.Errorf("class does not exist in registry")
	ErrDuplicatePath               = fmt.Errorf("duplicate path")
	ErrPathNotFound                = fmt.Errorf("path not found")
)

// Registry holds Classes and is responsible for making sure that
// every path within those classes is unique within the Registry.
// Once a class is registered with the registry, it will
// hook into the `Set` call of the class to ensure the integrity of the registry.
//
// TODO: Introduce the concept of 'generators' which allow to generate class files
// TODO: and manage them through skipper. This will remedy the missing 'Set' functionality.
type Registry struct {
	// classes map a classIdentifier string to the actual classes
	classes map[string]*Class
	// paths map absolute data paths to a classIdentifier string (key of the classes map)
	paths map[string]string
}

// NewRegistry returns a new, empty, registry.
func NewRegistry() *Registry {
	return &Registry{
		classes: make(map[string]*Class),
		paths:   make(map[string]string),
	}
}

type ClassLoaderFunc func(filePaths []string) ([]*Class, error)

func NewRegistryFromFiles(filePaths []string, classLoader ClassLoaderFunc) (*Registry, error) {
	classes, err := classLoader(filePaths)
	if err != nil {
		return nil, err
	}
	registry := NewRegistry()

	for _, class := range classes {
		err = registry.RegisterClass(class)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("unable to register class '%s'", class.Identifier), err)
		}
	}

	return registry, nil
}

func (reg *Registry) RegisterClass(class *Class) error {
	if _, exists := reg.classes[class.Identifier.String()]; exists {
		return fmt.Errorf("%s: %w", class.Identifier.String(), ErrClassAlreadyRegistered)
	}

	// Assemble the prefix to make all class relative paths regsitry-absolute
	// The last segment can safely be removed as we are certain that this is the class name
	// which is the same as the first segment of the paths emitted by the walk func
	pathPrefix := class.Identifier.StripSuffix(class.Identifier.LastSegment())

	// Create a slice of all absolute paths defined by the class
	var classPaths []string
	class.Walk(func(path data.Path, _ data.Value, _ bool) error {
		classPaths = append(classPaths, pathPrefix.AppendPath(path).String())
		return nil
	})

	// Now check whether any of the paths from the new class already exist.
	// This needs to be done even if it seems that the 'classIdentifier' prefix would be enough
	// separation.
	// There may be a classIdentifier 'foo' with a path 'foo.bar.baz'
	// and another classIdentifier 'foo.bar' with path 'foo.bar.baz' as well.
	// The purpose of the registry is to keep paths unique, hence we do need to verify that.
	var errs error
	for _, newClassPath := range classPaths {
		if _, exists := reg.paths[newClassPath]; exists {
			registeredByClass := reg.paths[newClassPath]
			errs = errors.Join(errs, fmt.Errorf("path already registered by '%s': %w: %s", registeredByClass, ErrDuplicatePath, newClassPath))
		}
	}
	if errs != nil {
		return errs
	}

	// inject hooks to monitor writes to the class which
	// ensure that the registry stays valid
	err := class.SetPreSetHook(reg.classPreSetHook())
	if err != nil {
		return fmt.Errorf("failed to register pre-set hook in class: %w", err)
	}
	err = class.SetPostSetHook(reg.classPostSetHook())
	if err != nil {
		return fmt.Errorf("failed to register post-set hook in class: %w", err)
	}

	// register class and all its paths
	for _, classPath := range classPaths {
		reg.paths[classPath] = class.Identifier.String()
	}
	reg.classes[class.Identifier.String()] = class

	return nil
}

// classPreSetHook is registered with every class which is part of the Registry.
// The hook makes sure that before a Class path is set, the action will not
// introduce any anomalies into the registry in order to keep its integrity
// and ensure that every path is uniquely pointing to just one value/class.
func (reg *Registry) classPreSetHook() SetHookFunc {
	return func(class Class, path data.Path, _ data.Value) error {
		registryPath := class.Identifier.StripSuffix(class.Identifier.LastSegment()).AppendPath(path)

		// if the path does already exist and is owned by a *different* class, then
		// we need to prevent the Set call as it would introduce a duplicate path.
		if classIdentifier, exists := reg.paths[registryPath.String()]; exists {
			existingPathClass, _ := reg.classes[classIdentifier]
			if !existingPathClass.Identifier.Equals(class.Identifier) {
				return fmt.Errorf("path already owned by '%s': %w: %s", classIdentifier, ErrDuplicatePath, registryPath)
			}
		}
		return nil
	}
}

// classPostSetHook is responsible for updating the registry in case
// the Set call on the Class introduced new path(s).
// Otherwise the registry would not know of the new paths and
// hence could not resolve the values at those paths.
func (reg *Registry) classPostSetHook() SetHookFunc {
	return func(class Class, path data.Path, value data.Value) error {
		registryPath := class.Identifier.StripSuffix(class.Identifier.LastSegment()).AppendPath(path)
		reg.paths[registryPath.String()] = class.Identifier.String()

		return nil
	}
}

// Get retrieves the value of a given Path if it exists.
func (reg *Registry) Get(path string) (data.Value, error) {
	if path == "" {
		return data.NilValue, data.ErrEmptyPath
	}

	// The path is expected to be registry-absolute (i.e. has a classIdentifier prefix).
	// Otherwise we cannot uniquely determine to which class the path belongs.
	// This is because for a given path 'baz', there may be
	// the classes 'foo' and 'foo.bar' in which the path 'baz' is valid.
	// Hence we cannot decide whether 'foo.baz' or 'foo.bar.baz' is meant.
	classIdentifier, exists := reg.paths[path]
	if exists {
		class := reg.classes[classIdentifier]
		classPath := data.NewPath(path).StripPrefix(data.NewPath(classIdentifier))
		return class.Get(classPath.String())
	}

	return data.NilValue, fmt.Errorf("path is not valid within the registry '%s': %w", path, ErrPathNotFound)
}

// GetPath is the same as Get, but it accepts a [data.Path] as path.
func (reg *Registry) GetPath(path data.Path) (data.Value, error) {
	return reg.Get(path.String())
}

// HasPath returns true if the path exists within the registry.
func (reg *Registry) HasPath(path data.Path) bool {
	_, err := reg.GetPath(path)
	if err != nil {
		return false
	}
	return true
}

// GetClassRelativePath attempts to resolve the target-class using the given classPath. Then it attempts to resolve the path (which might be class-local only).
// The given classPath can be any path which is known to the registry which is enough to resolve the Class.
func (reg *Registry) GetClassRelativePath(classPath data.Path, path data.Path) (data.Value, error) {
	classIdentifier, exists := reg.paths[classPath.String()]
	if !exists {
		return data.NilValue, fmt.Errorf("%s: %w", path, ErrPathNotFound)
	}

	class := reg.classes[classIdentifier]
	return class.GetPath(path)
}

// Set will set an existing (!) path within the registry.
// If the value is a complex type (map, array), then
// the newly created paths will be registered with the registry.
//
// Within the Registry one can only set paths which are already
// known by the registry. The reason for that is that the registry
// needs to resolve the class in which the path should be set.
// Assuming there exist two class identifiers 'foo' and 'foo.bar'.
// If the path 'foo.bar.baz' should be set, it cannot be determined
// in which class the path should be created, hence this functionality
// is only available within the Class itself.
func (reg *Registry) Set(path string, value interface{}) error {
	classIdentifier, exists := reg.paths[path]
	if !exists {
		return fmt.Errorf("cannot set unknown path: %s: %w", path, ErrPathNotFound)
	}
	return reg.classes[classIdentifier].Set(data.NewPath(path).StripPrefix(data.NewPath(classIdentifier)).String(), value)
}

// SetPath is just a wrapper for Set
func (reg *Registry) SetPath(path data.Path, value interface{}) error {
	return reg.Set(path.String(), value)
}

// WalkValues implements the [DataWalker] interface for the [Registry]
func (reg *Registry) Walk(walkFunc func(data.Path, data.Value, bool) error) error {
	for _, class := range reg.classes {
		err := class.Walk(func(path data.Path, value data.Value, isLeaf bool) error {
			if !isLeaf {
				return nil
			}
			return walkFunc(path, value, false)
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// GetClassByIdentifier attempts to return a class which is associated with the
// given classIdentifier.
// The classIdentifier cannot be empty.
// Returns an ErrClassIdentifierDoesNotExist if the identifier is unknown.
func (reg *Registry) GetClassByIdentifier(classIdentifier string) (*Class, error) {
	if classIdentifier == "" {
		return nil, ErrEmptyClassIdentifier
	}

	class, exists := reg.classes[classIdentifier]
	if !exists {
		return nil, ErrClassIdentifierDoesNotExist
	}
	return class, nil
}

// ClassIdentifiers returns a slice of all registered class identifiers as [data.Path]
func (reg *Registry) ClassIdentifiers() []data.Path {
	var ns []data.Path
	for namespace := range reg.classes {
		ns = append(ns, data.NewPath(namespace))
	}
	return ns
}

// AbsolutePath resolves the given, relative, path to an absolute path within the given context.
// This function satisfies the [skipper.AbsolutePathMaker] interface.
// The context is required to determine to which Class the path is relative to.
// The context can be any path within the class to which the path is relative to, even just the classIdentifier.
// In case the paths are empty or are not valid within the given context, an error is returned.
func (reg *Registry) AbsolutePath(path data.Path, context data.Path) (data.Path, error) {
	if path == nil || len(path) == 0 {
		return nil, data.ErrEmptyPath
	}
	if context == nil || len(context) == 0 {
		return nil, fmt.Errorf("context path cannot be empty: %w", data.ErrEmptyPath)
	}

	classIdentifier, exists := reg.paths[context.String()]
	if !exists {
		return nil, fmt.Errorf("unknown context path '%s': %w", context, ErrPathNotFound)
	}

	return reg.classes[classIdentifier].AbsolutePath(path, nil)
}
