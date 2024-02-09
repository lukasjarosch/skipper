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

	DataScope Scope = "data"
)

type Inventory struct {
	scopes       map[Scope]*Registry
	defaultScope Scope
}

func NewInventory() (*Inventory, error) {
	return &Inventory{
		scopes:       make(map[Scope]*Registry),
		defaultScope: "",
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

	inv.scopes[scope] = registry

	return nil
}

func (inv *Inventory) Get(path string) (data.Value, error) {
	// The path usually has the scope as the first segment
	for scope, registry := range inv.scopes {
		p := data.NewPath(path)
		if p.First() == string(scope) {
			return registry.Get(p.StripPrefix(p.FirstSegment()).String())
		}
	}

	// The path does not seem to have a scope prefix.
	// If there is a default scope, attempt to use that with the path.
	if inv.defaultScope != "" {
		return inv.scopes[inv.defaultScope].Get(path)
	}

	return data.NilValue, fmt.Errorf("%s: %w", path, ErrPathNotFound)
}

func (inv *Inventory) SetDefaultScope(scope Scope) error {
	if scope == "" {
		return ErrEmptyScope
	}

	if _, exists := inv.scopes[scope]; !exists {
		return fmt.Errorf("%s: %w", scope, ErrScopeDoesNotExist)
	}

	inv.defaultScope = scope

	return nil
}
