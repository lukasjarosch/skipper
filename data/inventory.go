package data

import (
	"fmt"
)

var (
	// RootNamespace is defined by an empty path
	RootNamespace = make(Path, 0)

	ErrContainerExists = fmt.Errorf("container already exists")
)

// Container defines an interface of a container holding values used by the [Inventory].
// This can be a regular [FileContainer], a dynamic container or any custom implementation.
type Container interface {
	Name() string
	AllPaths() []Path
	Get(Path) (interface{}, error)
}

type Inventory struct {
	// namespaces is a map of namespace strings to a map of container names -> [Container]
	namespaces map[string]map[string]Container
	// pathRegistry is a map of [Path]s strings to a [Scope] in which the value is located.
	pathRegistry map[string]ValueScope
}

func NewInventory() (*Inventory, error) {
	// create namespace registry and always register the root namespace
	ns := make(map[string]map[string]Container)
	ns[RootNamespace.String()] = make(map[string]Container)

	return &Inventory{
		namespaces:   ns,
		pathRegistry: make(map[string]ValueScope),
	}, nil
}

type RegisterOpts struct {
	ReplaceExisting bool
}
type RegisterOption func(*RegisterOpts)

func ReplaceContainer() RegisterOption {
	return func(ro *RegisterOpts) {
		ro.ReplaceExisting = true
	}
}

func (inv *Inventory) UnregisterContainer(namespace Path, containerName string) {
	if container, exists := inv.namespaces[namespace.String()][containerName]; exists {
		for _, path := range container.AllPaths() {
			delete(inv.pathRegistry, namespace.AppendPath(path).String())
		}

		delete(inv.namespaces[namespace.String()], containerName)
	}
}

func (inv *Inventory) RegisterContainer(namespace Path, container Container, options ...RegisterOption) error {
	// handle options
	opts := RegisterOpts{
		ReplaceExisting: false,
	}
	for _, opt := range options {
		opt(&opts)
	}

	if len(container.Name()) == 0 {
		return ErrEmptyContainerName
	}

	// just to be explicit about it, if the namespace is of length 0, it is always the root namespace
	if len(namespace) == 0 {
		namespace = RootNamespace
	}
	namespaceString := namespace.String()

	// ensure the namespace exists
	if _, exists := inv.namespaces[namespaceString]; !exists {
		inv.namespaces[namespaceString] = make(map[string]Container)
	}

	// If the container already exists unregister it if the ReplaceExisting option is set,
	// otherwise fail (default)
	if _, exists := inv.namespaces[namespaceString][container.Name()]; exists {
		if !opts.ReplaceExisting {
			return fmt.Errorf("%s: %w", container.Name(), ErrContainerExists)
		}
		inv.UnregisterContainer(namespace, container.Name())
	}

	// Value paths must stay unique within the [Inventory]
	// If one requests the value at 'foo.bar.baz', it cannot resolve to two different values.
	// For this we need to build up a path registry which allows us to quickly check for existing paths.
	{
		for _, path := range container.AllPaths() {
			// we're only interested in the absolute paths (namespace + path within container)
			checkPath := namespace.AppendPath(path)

			if scope, exists := inv.pathRegistry[checkPath.String()]; exists {
				return fmt.Errorf("path is already registered in namespace '%s' by container '%s': %s", scope.Namespace, scope.Container.Name(), checkPath)
			}

			inv.pathRegistry[checkPath.String()] = ValueScope{
				Namespace:     namespace,
				ContainerPath: path,
				Container:     container,
			}
		}
	}

	inv.namespaces[namespace.String()][container.Name()] = container

	return nil
}

func (inv *Inventory) GetValue(path Path) (Value, error) {
	scope, exists := inv.pathRegistry[path.String()]
	if !exists {
		return Value{}, fmt.Errorf("path does not exist: %s", path)
	}

	raw, err := scope.Container.Get(scope.ContainerPath)
	if err != nil {
		return Value{}, err
	}

	return Value{
		Raw:   raw,
		Scope: scope,
	}, nil
}

func (inv *Inventory) RegisteredNamespaces() []Path {
	namespaces := make([]Path, 0)
	for ns := range inv.namespaces {
		// the root namespace is always registered
		// is also empty, so there is not much use returning it here
		if ns == RootNamespace.String() {
			continue
		}

		namespaces = append(namespaces, NewPath(ns))
	}
	return namespaces
}

func (inv *Inventory) RegisteredPaths() []Path {
	var paths []Path
	for path := range inv.pathRegistry {
		paths = append(paths, NewPath(path))
	}
	return paths
}

func (val Value) String() string {
	return fmt.Sprint(val.Raw)
}
