package skipper

import "github.com/lukasjarosch/skipper/data"

// SetHookFunc can be registered as either preSetHook or postSetHook
// and will then be called respectively.
type SetHookFunc func(path data.Path, value data.Value) error

type HookablePostSet interface {
	RegisterPostSetHook(SetHookFunc)
}

type RegisterClassHookFunc func(class *Class) error

type HookablePostRegisterClass interface {
	RegisterPostRegisterClassHook(RegisterClassHookFunc)
}

type RegisterScopeHookFunc func(scope Scope, registry *Registry) error

type HookablePostRegisterScope interface {
	RegisterPostRegisterScopeHook(RegisterScopeHookFunc)
}
