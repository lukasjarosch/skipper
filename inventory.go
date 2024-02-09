package skipper

import (
	"fmt"

	"github.com/lukasjarosch/skipper/data"
)

type (
	// Namespace is a path which defines the scope of a class
	Namespace = data.Path
	// Identifier is a path which points to a value within a class.
	Identifier = data.Path
)

var (
	DefaultScope              = "default"
	ErrEmptyScope             = fmt.Errorf("empty scope")
	ErrClassAlreadyRegistered = fmt.Errorf("class already registered")
	ErrCannotResolvePath      = fmt.Errorf("unable to resolve path")
)

type Inventory struct {
	registry map[string]*Class
}

func NewInventory() (*Inventory, error) {
	return &Inventory{
		registry: make(map[string]*Class),
	}, nil
}

// RegisterClassWithScope will create a namespace <scope>.<classIdentifier> and attempt
// to register the class under it.
// Namespaces can only be registered once.
func (inv *Inventory) RegisterClassWithScope(scope string, classIdentifier data.Path, class *Class) error {
	if len(scope) == 0 {
		return ErrEmptyScope
	}

	// The namespace of a class is made up of the Inventory scope and the Identifier of the class.
	namespace := data.NewPath(scope)
	namespace = namespace.AppendPath(classIdentifier)

	// Every distinct namespace can only be registered once
	if _, exists := inv.registry[namespace.String()]; exists {
		return ErrClassAlreadyRegistered
	}

	inv.registry[namespace.String()] = class

	return nil
}

func (inv *Inventory) RegisterClass(classIdentifier data.Path, class *Class) error {
	return inv.RegisterClassWithScope(DefaultScope, classIdentifier, class)
}

func (inv *Inventory) Get(path string) (data.Value, error) {
	// Most likely the path is scoped to the default namespace, so let's just assume that.
	// Additionally, the default scope is the only scope which can be omitted.
	// To be sure attempt to remove the DefaultScope before re-adding it.
	searchPath := data.NewPath(path).StripPrefix(data.NewPath(DefaultScope)).Prepend(DefaultScope)

	// Now find out which namespace has the best match to the given path
	// This is easy, because the namespace needs to be the prefix of our search path
	for _, namespace := range inv.Namespaces() {
		// If the prefix matches, we've found the target and can resolve the value (hopefully)
		if searchPath.HasPrefix(namespace) {
			return inv.resolveValue(namespace, searchPath)
		}
	}

	// At this point the upper method failed us. Lets just try the full path.
	searchPath = data.NewPath(path)
	for _, namespace := range inv.Namespaces() {
		// If the prefix matches, we've found the target and can resolve the value (hopefully)
		if searchPath.HasPrefix(namespace) {
			return inv.resolveValue(namespace, searchPath)
		}
	}

	// Well, bummer...
	return data.NilValue, fmt.Errorf("%s: %w", path, ErrCannotResolvePath)
}

// Attempts to resolve a path within a given namespace.
//
// If the namespace is 'foo.bar' and the search path 'foo.bar.baz.qux'
// Then we can call Get with 'baz.qux' on the correct class.
// This works because the class name (baz) is always added to all paths by classes.
func (inv *Inventory) resolveValue(namespace data.Path, path data.Path) (data.Value, error) {
	valuePath := path.StripPrefix(namespace)
	class := inv.registry[namespace.String()]
	return class.Get(valuePath.String())
}

// Namespaces returns all registered namespaces
func (inv *Inventory) Namespaces() []Namespace {
	ns := []Namespace{}
	for namespace := range inv.registry {
		ns = append(ns, data.NewPath(namespace))
	}
	return ns
}
