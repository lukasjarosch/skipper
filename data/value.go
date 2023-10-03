package data

import "fmt"

// Value represents any value within the [Inventory]
// It has a [ValueScope] which determines the scope in which the value is defined.
type Value struct {
	Raw   interface{}
	Scope ValueScope
}

func (val Value) String() string {
	return fmt.Sprint(val.Raw)
}

// ValueScope defines a scope in which a value is valid / defined.
// It can also be used to retrieve the value again.
type ValueScope struct {
	// Namespace in which the value resides
	Namespace Path
	// The path within the container which points to the value
	ContainerPath Path
	// The actual container
	Container Container
}

// AbsolutePath returns the absolute path to this value.
// This is the namespace + ContainerPath.
// The AbsolutePath can be used to retrieve the value from the [Inventory]
//
// Note that the absolute path does not work with a scoped Inventory, only with the root Inventory.
// If you want to use it in a scoped inventory you need to strip the prefix yourself.
func (scope ValueScope) AbsolutePath() Path {
	return scope.Namespace.AppendPath(scope.ContainerPath)
}
