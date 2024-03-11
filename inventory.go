package skipper

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"

	"github.com/lukasjarosch/skipper/data"
)

// Scope defines a 'namespace' within the Inventory which is made up of a collection of classes.
type Scope string

var (
	DataScope    Scope = "data"
	TargetsScope Scope = "targets"
)

var (
	ErrEmptyScope                 = fmt.Errorf("scope is empty")
	ErrNilRegistry                = fmt.Errorf("registry is nil")
	ErrScopeDoesNotExist          = fmt.Errorf("scope does not exist")
	ErrScopeAlreadyRegistered     = fmt.Errorf("scope already registered")
	ErrTargetCannotIntroducePaths = fmt.Errorf("target cannot introduce new paths")
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

// Compile compiles the Inventory for a given target class.
//
// The target class is special as it is capable of overwriting values within the whole inventory.
// This is required because most of the time there will be values which are only valid per target.
// But in order to be able to model data wholistically, the user must be able to introduce
// paths which values only become defined once the target (say: environment) is defined.
//
// The target given must be a valid, inventory-absolute, class-path.
// So if the target to use originates from 'targets/develop.yaml', then
// the target path is 'targets.develop'.
//
// Within the target, one can overwrite every existing (!) path within the inventory.
// This works by iterating over every path of the target from which the class-name is stripped.
// If the remaining path is a valid inventory path, then the path is overwritten by the value defined in the target.
// Let's say that there exists a path 'data.common.network.name' with value 'AwesomeNet'.
// If the target class is 'develop' and it has a class-path 'develop.data.common.network.name' with value 'SuperNet',
// then the compiled inventory will have 'SuperNet' as value at 'data.common.network.name'.
//
// If within the target class, a non-existing
func (inv *Inventory) Compile(target data.Path) error {
	targetScope, err := inv.PathScope(target)
	if err != nil {
		return err
	}

	registry, err := inv.GetScope(targetScope)
	if err != nil {
		return err
	}

	// strip the scope prefix from the path, this should yield a valid classIdentifier within the registry
	classIdentifier := target.StripPrefix(data.NewPath(string(targetScope)))
	targetClass, err := registry.GetClassByIdentifier(classIdentifier.String())
	if err != nil {
		return err
	}

	// Overwrite inventory paths with target values if applicable.
	err = targetClass.WalkValues(func(p data.Path, v data.Value) error {
		pathWithoutClassName := p.StripPrefix(data.NewPath(targetClass.Name))

		// If the path does not start with a valid scope we simply ignore it.
		if _, err := inv.PathScope(pathWithoutClassName); err != nil {
			return nil
		}

		// If the path does not exist within the inventory, but the target attempted to set it
		// by using the scope prefix, abort.
		// This is not allowed, only known paths can be overwritten.
		// See [Registry.Set] for an explanation as of why.
		if _, err := inv.GetPath(pathWithoutClassName); err != nil {
			return fmt.Errorf("%w: path does not exist in inventory: %s", ErrTargetCannotIntroducePaths, pathWithoutClassName)
		}

		// The path exists within the inventory, overwrite it with the value from the target.
		spew.Println("OVERWRITTEN", pathWithoutClassName)
		err = inv.SetPath(pathWithoutClassName, v)
		if err != nil {
			return fmt.Errorf("failed to overwrite path %s: %w", pathWithoutClassName, err)
		}
		return nil
	})
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

// SetPath is a wrapper for [inventory.Set]
func (inv *Inventory) SetPath(path data.Path, value interface{}) error {
	return inv.Set(path.String(), value)
}

// Scopes returns all [Scope]s registered at the inventory
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
