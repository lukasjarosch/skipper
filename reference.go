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

var ErrValueReferenceSourceIsNil = fmt.Errorf("ValueReferenceSource cannot be nil")

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

	// register hooks to monitor the source
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

			// check if the reference is valid within the source,
			// otherwise there is no point in setting the value at all.
			_, err := manager.source.GetPath(newReference.AbsoluteTargetPath)
			if err != nil {
				return fmt.Errorf("invalid reference: %w", err)
			}

			// if the reference is already known and thus does not needed to be
			// added to the graph, continue
			if _, exists := manager.references[newReference.Hash()]; exists {
				continue
			}

			// TODO: remove
			reference.VisualizeDependencyGraph(manager.dependencyGraph, "/tmp/graph-pre.dot", "pre")

			// temporarily add the reference to the dependencyGraph to see whether
			// it will still be valid after the reference and it's dependencies are added.
			newReferenceDependencies := reference.ValueDependencies(newReference, manager.allReferences)
			err = reference.AddReferenceVertex(manager.dependencyGraph, newReference, newReferenceDependencies)
			if err != nil {
				return err
			}

			// TODO: remove
			reference.VisualizeDependencyGraph(manager.dependencyGraph, "/tmp/graph-post.dot", "post")

			err = reference.RemoveReferenceVertex(manager.dependencyGraph, newReference, newReferenceDependencies)
			if err != nil {
				return err
			}

			// TODO: remove
			reference.VisualizeDependencyGraph(manager.dependencyGraph, "/tmp/graph-post-delete.dot", "post-delete")

		}

		return nil
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

func (manager *ValueReferenceManager) postSetHook() SetHookFunc {
	return func(path data.Path, value data.Value) error {
		newReferences, err := reference.FindValueReference(manager.source, reference.ValueReferenceRegex, path, value)
		if err != nil {
			return err
		}

		existingReferences := manager.ReferencesAtPath(path)

		// Case 1: newReferences === existingReferences
		//
		// Case 2.1: newReferences !== existingReferences && len(newReferences) > len(existingReferences)
		// Case 2.2: newReferences !== existingReferences && len(newReferences) == len(existingReferences)
		// Case 2.3: newReferences !== existingReferences && len(newReferences) < len(existingReferences)
		//
		// For every case 2.x it might be easiest to just remove all existing references and add the newReferences

		_ = existingReferences

		// // No reference in value.
		// // This could still mean that there was a reference at the given path which is now being overwritten.
		// if len(references) == 0 {
		// 	var deleteIndices []int
		// 	for _, newReference := range references {
		// 		for i, reference := range manager.allReferences {
		// 			if reference.Path.Equals(newReference.Path) {
		// 				deleteIndices = append(deleteIndices, i)
		// 				break
		// 			}
		// 		}
		// 	}
		// }

		// TODO: what if the same reference was at that path multiple times and now is more/less times?
		// TODO: what if there were 3 references, but now are 2, or 4?

		for _, newReference := range newReferences {

			// check if the reference is still valid within the source,
			// otherwise there is no point in setting the value at all.
			_, err := manager.source.GetPath(newReference.AbsoluteTargetPath)
			if err != nil {
				return fmt.Errorf("invalid reference: %w", err)
			}

			// if the reference is already known it does not needed
			// to be resolved via the graph, but added to the list of all references.
			if _, exists := manager.references[newReference.Hash()]; exists {
				manager.allReferences = append(manager.allReferences, newReference)
				continue
			}

			// TODO: remove
			reference.VisualizeDependencyGraph(manager.dependencyGraph, "/tmp/graph-pre.dot", "pre")

			// temporarily add the reference to the dependencyGraph to see whether
			// it will still be valid after the reference and it's dependencies are added.
			newReferenceDependencies := reference.ValueDependencies(newReference, manager.allReferences)
			err = reference.AddReferenceVertex(manager.dependencyGraph, newReference, newReferenceDependencies)
			if err != nil {
				return err
			}

			// TODO: remove
			reference.VisualizeDependencyGraph(manager.dependencyGraph, "/tmp/graph-post.dot", "post")

			err = reference.RemoveReferenceVertex(manager.dependencyGraph, newReference, newReferenceDependencies)
			if err != nil {
				return err
			}

			// TODO: remove
			reference.VisualizeDependencyGraph(manager.dependencyGraph, "/tmp/graph-post-delete.dot", "post-delete")

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
