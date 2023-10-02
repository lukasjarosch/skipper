package data

import (
	"fmt"
)

var (
	// RootNamespace is defined by an empty path
	RootNamespace = make(Path, 0)

	ErrContainerExists    = fmt.Errorf("container already exists")
	ErrEmptyContainerName = fmt.Errorf("container name empty")
)

type ValueContainer interface {
	Name() string
	AllPaths() []Path
	Get(Path) (interface{}, error)
}

// ValueScope defines a scope in which a value is valid / defined.
type ValueScope struct {
	// Namespace in which the value resides
	Namespace Path
	// The path within the container which points to the value
	ContainerPath Path
	// The actual container
	Container ValueContainer
}

func (scope *ValueScope) AbsolutePath() Path {
	return scope.Namespace.AppendPath(scope.ContainerPath)
}

type Inventory struct {
	// namespaces is a map of namespace strings to a map of container names -> [Container]
	namespaces map[string]map[string]ValueContainer
	// pathRegistry is a map of [Path]s strings to a [Scope] in which the value is located.
	pathRegistry map[string]ValueScope
}

func NewInventory() (*Inventory, error) {
	// create namespace registry and always register the root namespace
	ns := make(map[string]map[string]ValueContainer)
	ns[RootNamespace.String()] = make(map[string]ValueContainer)

	return &Inventory{
		namespaces:   ns,
		pathRegistry: make(map[string]ValueScope),
	}, nil
}

func (inv *Inventory) RegisterContainer(namespace Path, container ValueContainer) error {
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
		inv.namespaces[namespaceString] = make(map[string]ValueContainer)
	}

	// existing containers are not overwritten
	if _, exists := inv.namespaces[namespaceString][container.Name()]; exists {
		return fmt.Errorf("%s: %w", container.Name(), ErrContainerExists)
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

type Value struct {
	Raw   interface{}
	Scope ValueScope
}

func (val Value) String() string {
	return fmt.Sprint(val.Raw)
}

func (inv *Inventory) GetValue(path Path) (Value, error) {
	scope, exists := inv.pathRegistry[path.String()]
	if !exists {
		return Value{}, fmt.Errorf("path does not exist: %s", path)
	}

	// remove the namespace from the path in order to query the container itself
	searchPath := path.StripPrefix(scope.Namespace)

	raw, err := scope.Container.Get(searchPath)
	if err != nil {
		return Value{}, err
	}

	return Value{
		Raw:   raw,
		Scope: scope,
	}, nil
}

// type ResolveResult struct {
// 	Namespace Path
// 	Container ValueContainer
// }
//
// func (inv *Inventory) ResolvePath(path Path) ([]ResolveResult, error) {
// 	log.Warn("attempting to resolve path", "path", path)
// 	log.SetLevel(log.DebugLevel)
//
// 	results := make([]ResolveResult, 0)
//
// 	for _, namespace := range inv.RegisteredNamespaces() {
// 		if !path.HasPrefix(namespace) {
// 			continue
// 		}
//
// 		// If the namespace matches, the next segment of the remaining path must
// 		// be a valid container name within that namespace
// 		//
// 		// e.g. If the path to resolve is 'foo.bar.baz' and the namespace 'foo' exists,
// 		// then there must be a container named 'bar' which can resolve 'bar.baz' or 'baz'
// 		// (should not matter as the container name must also be it's root key)
// 		remainingPath := path.StripPrefix(namespace)
// 		containerName := remainingPath.First()
//
// 		container, containerExists := inv.namespaces[namespace.String()][containerName]
// 		if !containerExists {
// 			continue
// 		}
//
// 		// TODO: what if the namespace 'foo.bar' exists and contains a container named 'baz'?
// 		// Is it valid at this point to resolve the whole container?
// 		// This is to be decided in the callee. If [Inventory.GetValue], then addressing
// 		// a whole container is invalid, if [Inventory.GetContainer], then addressing into containers may be invalid.
//
// 		if !container.HasPath(remainingPath) {
// 			continue
// 		}
//
// 		results = append(results, ResolveResult{
// 			Namespace: namespace,
// 			Container: container,
// 		})
// 	}
//
// 	// If the path 'foo.bar.baz' is to be resolved,
// 	// the [RootNamespace] could contain a [Container] named 'foo'
// 	// which can resolve the path `foo.bar.baz`.
// 	if container, exists := inv.namespaces[RootNamespace.String()][path.First()]; exists {
// 		log.Errorf("the root namespace has a container '%s'", path.First())
// 		_ = container
// 	}
//
// 	// TODO: consider the root namespace as candidate
//
// 	return nil, nil
// }

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
