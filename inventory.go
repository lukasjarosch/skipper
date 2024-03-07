package skipper

import (
	"fmt"

	"github.com/lukasjarosch/skipper/data"
)

type Scope string

var (
	ErrEmptyScope             = fmt.Errorf("scope is empty")
	ErrNilRegistry            = fmt.Errorf("registry is nil")
	ErrScopeDoesNotExist      = fmt.Errorf("scope does not exist")
	ErrScopeAlreadyRegistered = fmt.Errorf("scope already registered")

	DataScope    Scope = "data"
	TargetsScope Scope = "targets"
)

// Inventory is the top-level abstraction which represents all data.
// The Inventory is essentially just a wrapper over a map of Registries.
// It introduces the [Scope] which separates collections of Classes (Registries).
// Put simply, the Inventory is the projection of whatever is within the `inventory/` folder of a Skipper project.
type Inventory struct {
	scopes map[Scope]*Registry
	// hooks
	preRegisterScopeHooks  []RegisterScopeHookFunc
	postRegisterScopeHooks []RegisterScopeHookFunc
}

func NewInventory() (*Inventory, error) {
	return &Inventory{
		scopes: make(map[Scope]*Registry),
	}, nil
}

func (inv *Inventory) RegisterScope(scope Scope, registry *Registry) error {
	if scope == "" {
		return ErrEmptyScope
	}
	if registry == nil {
		return ErrNilRegistry
	}
	if _, exists := inv.scopes[scope]; exists {
		return fmt.Errorf("%s: %w", scope, ErrScopeAlreadyRegistered)
	}

	err := inv.callPreRegisterScopeHooks(scope, registry)
	if err != nil {
		return err
	}

	inv.scopes[scope] = registry

	err = inv.callPostRegisterScopeHooks(scope, registry)
	if err != nil {
		return err
	}

	return nil
}

// Get attempts to retrieve a [data.Value] from within the inventory.
// The path must begin with a valid [Scope], otherwise the Inventory
// will not be able to determine the correct [Registry] to search in.
func (inv *Inventory) Get(path string) (data.Value, error) {
	pathScope, err := inv.PathScope(data.NewPath(path))
	if err != nil {
		return data.NilValue, err
	}

	registry, err := inv.GetScope(pathScope)
	if err != nil {
		return data.NilValue, err
	}

	// remove the scope prefix as it is not valid within the registry
	registryPath := data.NewPath(path).StripPrefix(data.NewPath(string(pathScope)))

	return registry.Get(registryPath.String())
}

func (inv *Inventory) GetPath(path data.Path) (data.Value, error) {
	return inv.Get(path.String())
}

func (inv *Inventory) GetScope(scope Scope) (*Registry, error) {
	if scope == "" {
		return nil, ErrEmptyScope
	}
	registry, exists := inv.scopes[scope]
	if !exists {
		return nil, ErrScopeDoesNotExist
	}

	return registry, nil
}

func (inv *Inventory) GetClassRelativePath(scopedClassPath data.Path, path data.Path) (data.Value, error) {
	scope, err := inv.PathScope(scopedClassPath)
	if err != nil {
		return data.NilValue, err
	}

	registry, err := inv.GetScope(scope)
	if err != nil {
		return data.NilValue, err
	}

	classPath := scopedClassPath.StripPrefix(data.NewPath(string(scope)))

	return registry.GetClassRelativePath(classPath, path)
}

// Set will set an existing (!) path within the inventory.
// It is assumed that the path has a scope prefix.
// So initially every scope will be checked for a match
func (inv *Inventory) Set(path string, value interface{}) error {
	pathScope, err := inv.PathScope(data.NewPath(path))
	if err != nil {
		return err
	}

	registry, err := inv.GetScope(pathScope)
	if err != nil {
		return err
	}

	// remove the scope prefix as it is not valid within the registry
	registryPath := data.NewPath(path).StripPrefix(data.NewPath(string(pathScope)))

	return registry.Set(registryPath.String(), value)
}

func (inv *Inventory) SetPath(path data.Path, value interface{}) error {
	return inv.Set(path.String(), value)
}

func (inv *Inventory) Scopes() []Scope {
	var scopes []Scope
	for scope := range inv.scopes {
		scopes = append(scopes, scope)
	}
	return scopes
}

// PathScope returns the scope in which the path is defined.
// If the given path is not inventory-absolute (i.e. has a scope prefix) an error is returned.
// This is because a path without scope prefix (e.g. 'foo.bar') may be valid in multiple scopes.
func (inv *Inventory) PathScope(path data.Path) (Scope, error) {
	firstSegment := path.First()

	if _, exists := inv.scopes[Scope(firstSegment)]; exists {
		return Scope(firstSegment), nil
	}

	return Scope(""), fmt.Errorf("path is not valid in any scope '%s': %w", path, ErrPathNotFound)
}

// Walk allows to use depth-first-search to walk over all paths which point to scalars (leaf nodes).
func (inv *Inventory) WalkValues(walkFunc func(data.Path, data.Value) error) error {
	for scope, registry := range inv.scopes {
		err := registry.WalkValues(func(path data.Path, value data.Value) error {
			return walkFunc(path.Prepend(string(scope)), value)
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// AbsolutePath resolves the given, relative, path to an absolute path within the given context.
// This function satisfies the [skipper.AbsolutePathMaker] interface.
// The context is required to determine to which Class the path is relative to.
// The context can be any path within the class to which the path is relative to.
// In case the paths are empty or are not valid within the given context, an error is returned.
func (inv *Inventory) AbsolutePath(path data.Path, context data.Path) (data.Path, error) {
	if len(path) == 0 {
		return nil, data.ErrEmptyPath
	}
	if len(context) == 0 {
		return nil, fmt.Errorf("context path cannot be empty: %w", data.ErrEmptyPath)
	}

	// maybe the path is already a valid absolute path?
	_, err := inv.GetPath(path)
	if err == nil { // Note the err == nil!
		return path, nil
	}

	// the context must be an absolute path, hence we should be able to get the scope easily
	scope, err := inv.PathScope(context)
	if err != nil {
		return nil, err
	}

	// let the registry resolve the rest of the context path
	abs, err := inv.scopes[scope].AbsolutePath(path, context.StripPrefix(data.NewPath(string(scope))))
	if err != nil {
		return nil, err
	}

	// make sure the scope is the prefix
	if !abs.HasPrefix(data.NewPath(string(scope))) {
		abs = abs.Prepend(string(scope))
	}

	return abs, nil
}

func (inv *Inventory) Values() map[string]data.Value {
	valueMap := make(map[string]data.Value)

	_ = inv.WalkValues(func(p data.Path, v data.Value) error {
		valueMap[p.String()] = v
		return nil
	})

	return valueMap
}

// RegisterPreSetHook implements the [HookableSet] interface.
// Because each [Class] also implements that interface, this
// func is just going to redirect the call to every class within each registry.
func (inv *Inventory) RegisterPreSetHook(hook SetHookFunc) {
	for scope, reg := range inv.scopes {
		for _, class := range reg.ClassMap() {
			class.RegisterPreSetHook(func(path data.Path, value data.Value) error {
				// make sure the path contains the scope
				path = path.Prepend(string(scope))
				return hook(path, value)
			})
		}
	}
}

// RegisterPostSetHook implements the [HookableSet] interface.
// Because each [Class] also implements that interface, this
// func is just going to redirect the call to every class within each registry.
func (inv *Inventory) RegisterPostSetHook(hook SetHookFunc) {
	for scope, reg := range inv.scopes {
		for _, class := range reg.ClassMap() {
			class.RegisterPostSetHook(func(path data.Path, value data.Value) error {
				// make sure the path contains the scope
				path = path.Prepend(string(scope))
				return hook(path, value)
			})
		}
	}
}

func (inv *Inventory) RegisterPreRegisterScopeHook(hook RegisterScopeHookFunc) {
	inv.preRegisterScopeHooks = append(inv.preRegisterScopeHooks, hook)
}

func (inv *Inventory) RegisterPostRegisterScopeHook(hook RegisterScopeHookFunc) {
	inv.postRegisterScopeHooks = append(inv.postRegisterScopeHooks, hook)
}

func (inv *Inventory) callPreRegisterScopeHooks(scope Scope, registry *Registry) error {
	for _, hook := range inv.preRegisterScopeHooks {
		err := hook(scope, registry)
		if err != nil {
			return err
		}
	}
	return nil
}

func (inv *Inventory) callPostRegisterScopeHooks(scope Scope, registry *Registry) error {
	for _, hook := range inv.postRegisterScopeHooks {
		err := hook(scope, registry)
		if err != nil {
			return err
		}
	}
	return nil
}
