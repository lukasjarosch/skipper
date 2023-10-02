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

type Scope struct {
	Namespace Path
	Container *Container
}

type Inventory struct {
	// namespaces is a map of namespace strings to a map of container names -> [Container]
	namespaces map[string]map[string]*Container
	// pathRegistry is a map of [Path]s strings to a bool (true) used to quickly check the existence of a path
	pathRegistry map[string]Scope
}

func NewInventory() (*Inventory, error) {
	// create namespace registry and always register the root namespace
	ns := make(map[string]map[string]*Container)
	ns[RootNamespace.String()] = make(map[string]*Container)

	return &Inventory{
		namespaces:   ns,
		pathRegistry: make(map[string]Scope),
	}, nil
}

func (inv *Inventory) RegisterContainer(namespace Path, container *Container) error {
	if len(container.Name) == 0 {
		return ErrEmptyContainerName
	}

	// just to be explicit about it, if the namespace is of length 0, it is always the root namespace
	if len(namespace) == 0 {
		namespace = RootNamespace
	}
	namespaceString := namespace.String()

	// ensure the namespace exists
	if _, exists := inv.namespaces[namespaceString]; !exists {
		inv.namespaces[namespaceString] = make(map[string]*Container)
	}

	// existing containers are not overwritten
	if _, exists := inv.namespaces[namespaceString][container.Name]; exists {
		return fmt.Errorf("%s: %w", container.Name, ErrContainerExists)
	}

	// Value paths must stay unique within the [Inventory]
	// If one requests the value at 'foo.bar.baz', it cannot resolve to two different values.
	// For this we need to build up a path registry which allows us to quickly check for existing paths.
	{
		for _, path := range container.AllPaths() {
			// we're only interested in the absolute paths (namespace + path within container)
			checkPath := namespace.AppendPath(path)

			if scope, exists := inv.pathRegistry[checkPath.String()]; exists {
				return fmt.Errorf("path is already registered in namespace '%s' by container '%s': %s", scope.Namespace, scope.Container.Name, checkPath)
			}

			inv.pathRegistry[checkPath.String()] = Scope{
				Namespace: namespace,
				Container: container,
			}
		}
	}

	inv.namespaces[namespace.String()][container.Name] = container

	return nil
}
