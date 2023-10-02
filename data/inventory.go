package data

import (
	"fmt"
)

var (
	// RootNamespace is defined by an empty path
	RootNamespace = make(Path, 0)

	ErrContainerExists = fmt.Errorf("container already exists")
	ErrNilContainer    = fmt.Errorf("container is nil")
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
	// pathRegistry is a map of [Path]s strings to a [ValueScope] in which the value is located.
	pathRegistry map[string]ValueScope
	// prefix all paths with this, if len==0 this is disabled
	prefix Path
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

func (inv *Inventory) Scoped(prefix Path) *Inventory {
	return &Inventory{
		namespaces:   inv.namespaces,
		pathRegistry: inv.pathRegistry,
		prefix:       prefix,
	}
}

type RegisterOpts struct {
	ReplaceExisting bool
	Patch           bool
}
type RegisterOption func(*RegisterOpts)

func ReplaceContainer() RegisterOption {
	return func(ro *RegisterOpts) {
		ro.ReplaceExisting = true
	}
}

func Patch() RegisterOption {
	return func(ro *RegisterOpts) {
		ro.Patch = true
	}
}

func (inv *Inventory) prefixed(path Path) Path {
	if len(inv.prefix) == 0 {
		return path
	}
	return inv.prefix.AppendPath(path)
}

func (inv *Inventory) UnregisterContainer(namespace Path, containerName string) {
	namespace = inv.prefixed(namespace)

	if container, exists := inv.namespaces[namespace.String()][containerName]; exists {
		for _, path := range container.AllPaths() {
			delete(inv.pathRegistry, namespace.AppendPath(path).String())
		}

		delete(inv.namespaces[namespace.String()], containerName)
	}
}

func (inv *Inventory) RegisterContainer(namespace Path, container Container, options ...RegisterOption) error {
	namespace = inv.prefixed(namespace)

	if container == nil {
		return ErrNilContainer
	}
	if len(container.Name()) == 0 {
		return ErrEmptyContainerName
	}

	// handle options
	opts := RegisterOpts{
		ReplaceExisting: false,
		Patch:           false,
	}
	for _, opt := range options {
		opt(&opts)
	}

	// just to be explicit about it, if the namespace is of length 0, it is always the root namespace
	if len(namespace) == 0 {
		namespace = inv.prefixed(RootNamespace)
	}
	namespaceString := namespace.String()

	// ensure the namespace exists
	if _, exists := inv.namespaces[namespaceString]; !exists {
		inv.namespaces[namespaceString] = make(map[string]Container)
	}

	// If the container already exists unregister it if the ReplaceExisting option is set,
	// otherwise fail (default)
	if _, exists := inv.namespaces[namespaceString][container.Name()]; exists {
		if !opts.ReplaceExisting && !opts.Patch {
			return fmt.Errorf("%s.%s: %w", namespace, container.Name(), ErrContainerExists)
		}
		// if replace mode is enabled, remove the existing container and it's data alltogether
		if opts.ReplaceExisting {
			inv.UnregisterContainer(namespace, container.Name())
		}
	}

	// Value paths must stay unique within the [Inventory]
	// If one requests the value at 'foo.bar.baz', it cannot resolve to two different values.
	// For this we need to build up a path registry which allows us to quickly check for existing paths.
	{
		for _, path := range container.AllPaths() {
			// we're only interested in the absolute paths (namespace + path within container)
			absPath := inv.prefixed(namespace).AppendPath(path)

			if scope, exists := inv.pathRegistry[absPath.String()]; exists {
				if !opts.Patch {
					return fmt.Errorf("path is already registered in namespace '%s' by container '%s': %s", inv.prefixed(scope.Namespace), scope.Container.Name(), absPath)
				}

				// the path exists, but the patch option is set
				// in this case we will continue and allow the scope to be changed to point to the patch container.
				// WARN: the patch container is not in the namespace map, is this an issue? (One cannot directly fetch that container easily)
				// TODO: preserve original scope to keep access to the original data
			}

			inv.pathRegistry[absPath.String()] = ValueScope{
				Namespace:     namespace,
				ContainerPath: path,
				Container:     container,
			}
		}
	}

	inv.namespaces[namespace.String()][container.Name()] = container

	return nil
}

func (inv *Inventory) MustGetValue(path Path) Value {
	val, err := inv.GetValue(path)
	if err != nil {
		panic(err)
	}
	return val
}

func (inv *Inventory) GetValue(path Path) (Value, error) {
	path = inv.prefixed(path)

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
		nsp := NewPath(ns)

		// if we are in a scoped inventory all namespaces without the prefix
		// must be removed and those with the prefix must have it removed
		if len(inv.prefix) > 0 {
			if !nsp.HasPrefix(inv.prefix) {
				continue
			}
			nsp = nsp.StripPrefix(inv.prefix)
		}

		// the root namespace is always registered
		// is also empty, so there is not much use returning it here
		if nsp.String() == RootNamespace.String() {
			continue
		}

		namespaces = append(namespaces, nsp)
	}

	SortPaths(namespaces)

	return namespaces
}

func (inv *Inventory) RegisteredPaths() []Path {
	var paths []Path
	for path := range inv.pathRegistry {
		ppath := NewPath(path)

		// if we are in a scoped inventory all paths without the prefix
		// must be removed and those with the prefix must have it removed
		if len(inv.prefix) > 0 {
			if !ppath.HasPrefix(inv.prefix) {
				continue
			}
			ppath = ppath.StripPrefix(inv.prefix)
		}

		paths = append(paths, ppath)
	}

	SortPaths(paths)

	return paths
}

func (val Value) String() string {
	return fmt.Sprint(val.Raw)
}
