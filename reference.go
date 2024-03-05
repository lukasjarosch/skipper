package skipper

import (
	"fmt"

	"github.com/dominikbraun/graph"

	"github.com/lukasjarosch/skipper/data"
	"github.com/lukasjarosch/skipper/reference"
)

type ValueReferenceSource interface {
	HookableSet
	reference.ValueSource
	reference.ValueTarget
}

var (
	ErrValueReferenceSourceIsNil  = fmt.Errorf("ValueReferenceSource cannot be nil")
	ErrReferenceDoesNotExist      = fmt.Errorf("reference does not exist")
	ErrInvalidReferencePath       = fmt.Errorf("invalid reference path")
	ErrInvalidReferenceTargetPath = fmt.Errorf("invalid reference target path")
	ErrInvalidReferenceNotInValue = fmt.Errorf("invalid reference, no reference in value")
)

type ValueReferenceManager struct {
	source ValueReferenceSource
	// allReferences contains all found references, even duplicates
	allReferences []reference.ValueReference
	// references maps reference hashes to references, hence being
	// the deduplicated version of allReferences
	references map[string]reference.ValueReference
	// stores all references and their dependencies
	dependencyGraph graph.Graph[string, reference.ValueReference]
}

func NewValueReferenceManager(source ValueReferenceSource) (*ValueReferenceManager, error) {
	if source == nil {
		return nil, ErrValueReferenceSourceIsNil
	}

	manager := &ValueReferenceManager{
		source:     source,
		references: make(map[string]reference.ValueReference),
	}

	// parse out all value references
	references, err := reference.FindAllValueReferences(manager.source)
	if err != nil {
		return nil, err
	}
	manager.allReferences = references

	// deduplicate references in map
	for _, ref := range manager.allReferences {
		manager.references[ref.Hash()] = ref
	}

	manager.registerHooks()

	// create dependency graph
	manager.dependencyGraph, err = reference.ValueDependencyGraph(references)
	if err != nil {
		return nil, err
	}

	return manager, nil
}

func (manager ValueReferenceManager) AllReferences() []reference.ValueReference {
	return manager.allReferences
}

func (manager ValueReferenceManager) ReferenceMap() map[string]reference.ValueReference {
	return manager.references
}

func (manager *ValueReferenceManager) registerHooks() {
	// We need to be aware of every 'write' (Set) operation
	// within the source.
	manager.source.RegisterPreSetHook(manager.preSetHook())
	manager.source.RegisterPostSetHook(manager.postSetHook())

	// In case we're dealing with a Registry, we also need to
	// track any new classes being added to it.
	if regSource, ok := manager.source.(HookableRegisterClass); ok {
		regSource.RegisterPreRegisterClassHook(manager.preRegisterClassHook())
		regSource.RegisterPostRegisterClassHook(manager.postRegisterClassHook())
	}

	// In case we're dealing with an Inventory, we also need to
	// track any new scopes being added.
	if invSource, ok := manager.source.(HookableRegisterScope); ok {
		invSource.RegisterPreRegisterScopeHook(manager.preRegisterScopeHook())
		invSource.RegisterPostRegisterScopeHook(manager.postRegisterScopeHook())
	}
}

func (manager *ValueReferenceManager) ReferencesAtPath(path data.Path) []reference.ValueReference {
	var refs []reference.ValueReference
	for _, ref := range manager.allReferences {
		if ref.Path.Equals(path) {
			refs = append(refs, ref)
		}
	}
	return refs
}

// ValidateReference checks whether the given reference is valid within the manager's source.
// It will check that the 'Path' and 'AbsoluteTargetPath' exist.
// It will also parse the value at 'Path' and ensure that the given reference exists within that value.
// Note: Because of this, do not use this function before a value has been set (postSet)!
func (manager *ValueReferenceManager) ValidateReference(ref reference.ValueReference) error {
	if _, err := manager.source.GetPath(ref.Path); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidReferencePath, err)
	}
	if _, err := manager.source.GetPath(ref.AbsoluteTargetPath); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidReferenceTargetPath, err)
	}

	refPathValue, _ := manager.source.GetPath(ref.Path)
	foundRefs, err := reference.FindValueReference(manager.source, reference.ValueReferenceRegex, ref.Path, refPathValue)
	if err != nil {
		return err
	}

	// at that path, there may exist multiple references, but at least the reference which is to be added must exist.
	foundRef := false
	for _, found := range foundRefs {
		if found.Hash() == ref.Hash() {
			foundRef = true
			break
		}
	}
	if !foundRef {
		return fmt.Errorf("reference '%s' does not exist at path %s: %w", ref.Name(), ref.Path, ErrInvalidReferenceNotInValue)
	}

	return nil
}

// addReference adds a new [reference.ValueReference] to the manager.
// It takes care that new, unique, references are also added to the hash-map and the graph.
func (manager *ValueReferenceManager) addReference(ref reference.ValueReference) error {
	err := manager.ValidateReference(ref)
	if err != nil {
		return fmt.Errorf("invalid reference: %w", err)
	}

	manager.allReferences = append(manager.allReferences, ref)

	// If the new reference does not introduce a new hash, nothing needs to be resolved.
	// We can add the reference to the list of known references and be done.
	if _, exists := manager.references[ref.Hash()]; exists {
		return nil
	}

	// Otherwise, register the hash, resolve the dependencies and update the graph
	manager.references[ref.Hash()] = ref
	dependencies := reference.ValueDependencies(ref, manager.allReferences)
	err = reference.AddReferenceVertex(manager.dependencyGraph, ref, dependencies)
	if err != nil {
		return err
	}

	return nil
}

// removeReference removes one instance of the given reference.
// If after the removal, no other instances of the same reference are known by the manager,
// it is forgotten about completely.
func (manager *ValueReferenceManager) removeReference(ref reference.ValueReference) error {
	if _, exists := manager.references[ref.Hash()]; !exists {
		return ErrReferenceDoesNotExist
	}

	// find any instance of the reference and safe its index
	var removeIndex int
	for i, existingRef := range manager.AllReferences() {
		if existingRef.Hash() == ref.Hash() {
			removeIndex = i
			break
		}
	}

	// forget about the reference at 'removeIndex' by overwriting it with the last reference in the slice
	manager.allReferences[removeIndex] = manager.allReferences[len(manager.allReferences)-1]
	manager.allReferences = manager.allReferences[:len(manager.allReferences)-2]

	// determine if there are any instances of the reference left
	var forgetReference bool
	for _, existingRef := range manager.AllReferences() {
		if existingRef.Hash() == ref.Hash() {
			forgetReference = false
		}
	}

	if !forgetReference {
		return nil
	}

	// remove reference from graph
	err := reference.RemoveReferenceVertex(manager.dependencyGraph, ref)
	if err != nil {
		return fmt.Errorf("failed to remove reference from graph: %w", err)
	}

	// forget about the reference
	delete(manager.references, ref.Hash())

	return nil
}

func (manager *ValueReferenceManager) preSetHook() SetHookFunc {
	return func(path data.Path, value data.Value) error {
		references, err := reference.FindValueReference(manager.source, reference.ValueReferenceRegex, path, value)
		if err != nil {
			return err
		}

		// If there is no reference in the value, there is nothing to do before setting the value.
		// It could still be that there was a reference at that path which needs to be removed.
		// That case is handled in postSetHook, though.
		if len(references) == 0 {
			return nil
		}

		for _, newReference := range references {
			if _, err := manager.source.GetPath(newReference.AbsoluteTargetPath); err != nil {
				return fmt.Errorf("%w: %w", ErrInvalidReferenceTargetPath, err)
			}

			// continue, if the reference is already known and hence does not needed to be added to the graph
			if _, exists := manager.references[newReference.Hash()]; exists {
				continue
			}

			// temporarily add the reference to the dependencyGraph to see whether
			// it will still be valid after the reference and it's dependencies are added.
			newReferenceDependencies := reference.ValueDependencies(newReference, manager.allReferences)
			err = reference.AddReferenceVertex(manager.dependencyGraph, newReference, newReferenceDependencies)
			if err != nil {
				return err
			}

			// nice, now remove it from the graph again
			err = reference.RemoveReferenceVertex(manager.dependencyGraph, newReference)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

func (manager *ValueReferenceManager) postSetHook() SetHookFunc {
	return func(path data.Path, value data.Value) error {
		newReferences, err := reference.FindValueReference(manager.source, reference.ValueReferenceRegex, path, value)
		if err != nil {
			return err
		}

		// Instead of figuring out which reference was added or removed,
		// we simply remove all existing references and add the new ones.
		for _, existingRef := range manager.ReferencesAtPath(path) {
			err = manager.removeReference(existingRef)
			if err != nil {
				return fmt.Errorf("failed to remove existing reference: %w", err)
			}
		}
		for _, newRef := range newReferences {
			err = manager.addReference(newRef)
			if err != nil {
				return fmt.Errorf("failed to add new reference: %w", err)
			}
		}

		return nil
	}
}

func (manager *ValueReferenceManager) preRegisterScopeHook() RegisterScopeHookFunc {
	return func(scope Scope, registry *Registry) error {
		return nil
	}
}

func (manager *ValueReferenceManager) postRegisterScopeHook() RegisterScopeHookFunc {
	return func(scope Scope, registry *Registry) error {
		return nil
	}
}

func (manager *ValueReferenceManager) preRegisterClassHook() RegisterClassHookFunc {
	return func(class *Class) error {
		return nil
	}
}

func (manager *ValueReferenceManager) postRegisterClassHook() RegisterClassHookFunc {
	return func(class *Class) error {
		return nil
	}
}
