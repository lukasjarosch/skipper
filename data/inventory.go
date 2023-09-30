package data

import "fmt"

var (
	// RootNamespace is defined by an empty path
	RootNamespace = make(Path, 0)

	ErrContainerExists    = fmt.Errorf("container already exists")
	ErrEmptyContainerName = fmt.Errorf("container name empty")
)

type Inventory struct {
	namespaces map[string]map[string]*Container
}

func NewInventory() (*Inventory, error) {
	return &Inventory{
		namespaces: make(map[string]map[string]*Container, 0),
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

	inv.namespaces[namespace.String()][container.Name] = container

	return nil
}
