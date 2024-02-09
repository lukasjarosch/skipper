package skipper

import (
	"errors"
	"fmt"

	"github.com/davecgh/go-spew/spew"

	"github.com/lukasjarosch/skipper/data"
)

var (
	ErrNamespaceDoesNotExist  = fmt.Errorf("namespace does not exist")
	ErrEmptyClassIdentifier   = fmt.Errorf("class identifier cannot be empty")
	ErrClassAlreadyRegistered = fmt.Errorf("class already registered in namespace")
	ErrClassDoesNotExist      = fmt.Errorf("class does not exist in registry")
	ErrInvalidClassIdentifier = fmt.Errorf("invalid class identifier")
	ErrDuplicatePath          = fmt.Errorf("duplicate path")
)

// Registry holds Classes and is responsible for making sure that
// every path within those classes is unique within the Registry.
type Registry struct {
	// classes map a classIdentifier string to the actual classes
	classes map[string]*Class
	// paths map absolute data paths to a classIdentifier string (key of the classes map)
	paths map[string]string
}

func NewRegistry() *Registry {
	return &Registry{
		classes: make(map[string]*Class),
		paths:   make(map[string]string),
	}
}

func (reg *Registry) RegisterClass(classIdentifier data.Path, class *Class) error {
	if classIdentifier.String() == "" {
		return ErrEmptyClassIdentifier
	}
	if classIdentifier.Last() != class.Name {
		return fmt.Errorf("class name must be last segment: %w", ErrInvalidClassIdentifier)
	}
	if _, exists := reg.classes[classIdentifier.String()]; exists {
		return fmt.Errorf("%s: %w", classIdentifier.String(), ErrClassAlreadyRegistered)
	}

	// Assemble the prefix to make all class relative paths regsitry-absolute
	// The last segment can safely be removed as we are certain that this is the class name
	// which is the same as the first segment of the paths emitted by the walk func
	pathPrefix := classIdentifier.StripSuffix(classIdentifier.LastSegment())

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
			registeredByClass := reg.classes[reg.paths[newClassPath]].Name
			errs = errors.Join(errs, fmt.Errorf("path already registered by '%s': %w: %s", registeredByClass, ErrDuplicatePath, newClassPath))
		}
	}
	if errs != nil {
		return errs
	}

	err := class.SetPreSetHook(func(class Class, path data.Path, _ data.Value) error {
		classIdentifierStr, err := reg.GetClassIdentifierByClassId(class.ID())
		if err != nil {
			return fmt.Errorf("unable to locate identifier of the class, this should not happen")
		}

		// strip the class name from the identifier and assemble registry-absolute path
		// which would be created by this 'Set' call
		classIdentifier := data.NewPath(classIdentifierStr)
		registryPath := classIdentifier.StripSuffix(classIdentifier.LastSegment()).AppendPath(path)

		// if the path does already exist and is owned by a *different* class, then
		// we need to prevent the Set call as it would introduce a duplicate path.
		if classIdentifier, exists := reg.paths[registryPath.String()]; exists {
			existingPathClass, _ := reg.classes[classIdentifier]
			if existingPathClass.ID() != class.ID() {
				return fmt.Errorf("path already owned by '%s': %w: %s", classIdentifier, ErrDuplicatePath, registryPath)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to register pre-set hook in class: %w", err)
	}

	err = class.SetPostSetHook(func(class Class, path data.Path, _ data.Value) error {
		spew.Println("POST WRITE", path)
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to register post-set hook in class: %w", err)
	}

	reg.classes[classIdentifier.String()] = class
	for _, classPath := range classPaths {
		reg.paths[classPath] = classIdentifier.String()
	}

	return nil
}

func (reg *Registry) GetClassByIdentifier(classIdentifier string) (*Class, error) {
	if classIdentifier == "" {
		return nil, ErrEmptyClassIdentifier
	}

	class, exists := reg.classes[classIdentifier]
	if !exists {
		return nil, ErrNamespaceDoesNotExist
	}
	return class, nil
}

func (reg *Registry) GetClassIdentifierByClassId(id string) (string, error) {
	if id == "" {
		return "", ErrEmptyClassId
	}

	for identifier, class := range reg.classes {
		if class.ID() == id {
			return identifier, nil
		}
	}

	return "", fmt.Errorf("%w: %s", ErrClassDoesNotExist, id)
}

func (reg *Registry) ResolveClass(path data.Path) (*Class, error) {
	// If the path is 'foo.bar.baz' it could map to:
	//   namespace 'foo'     -> class 'bar' -> path 'baz'
	//   namespace '[root]'  -> class 'foo' -> path 'bar.baz'
	//   namespace 'foo.bar' -> class 'baz'

	for _, ns := range reg.ClassIdentifiers() {
		spew.Dump(ns)
	}
	return nil, nil
}

func (reg *Registry) ClassIdentifiers() []data.Path {
	var ns []data.Path
	for namespace := range reg.classes {
		ns = append(ns, data.NewPath(namespace))
	}
	return ns
}
