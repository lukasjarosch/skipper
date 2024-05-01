package skipper

import (
	"fmt"

	"github.com/dominikbraun/graph"

	"github.com/lukasjarosch/skipper/v1/data"
	"github.com/lukasjarosch/skipper/v1/reference"
)

type ValueReferenceSource interface {
	HookablePostSet
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

// ValueReferenceManager is responsible for managing all [reference.ValueReference]
// within the given [ValueReferenceSource].
// It leverages hooks to monitor any runtime changes of the source to keep a
// consistent list of all references within it.
// The ValueReferenceManager is also able to perform a replace of all
// references within the source.
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

// NewValueReferenceManager constructs a new [ValueReferenceManager] with the given [ValueReferenceSource].
// The source is completely parsed for [reference.ValueReference]s and an initial dependency-graph is constructed.
// Additionally, all hooks of the manager are registered with the source.
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

	// make sure the references are valid
	for _, newRef := range references {
		err := manager.ValidateReference(newRef)
		if err != nil {
			return nil, err
		}
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

func (manager *ValueReferenceManager) getAllReferencesWithHash(hash string) []reference.ValueReference {
	var result []reference.ValueReference
	for _, ref := range manager.AllReferences() {
		if ref.Hash() == hash {
			result = append(result, ref)
		}
	}
	return result
}

// ReplaceReferences will replace all - currently known - references
// within the source of the manager.
// If any references are added/removed after this call, this
// method needs to be called again.
//
// This method is idempotent.
func (manager *ValueReferenceManager) ReplaceReferences() error {
	uniqueReplacementOrder, err := reference.ValueReplacementOrder(manager.dependencyGraph)
	if err != nil {
		return err
	}

	// The replacementOrder is derived from the hash-map, hence it does
	// not know about duplicate references.
	// Calculate the actual replacement order including the duplicates.
	var replacementOrder []reference.ValueReference
	for _, orderedRef := range uniqueReplacementOrder {
		replacementOrder = append(replacementOrder, manager.getAllReferencesWithHash(orderedRef.Hash())...)
	}

	return reference.ReplaceValues(manager.source, replacementOrder)
}

// registerHooks is responsible for registering all possible hooks which
// this manager provides.
// It will always register the 'preSetHook' and 'postSetHook' functions
// to the source.
// If the source implements the [HookableRegisterClass] interface,
// the 'postRegisterClassHook' is registered.
// If the source implements the [HookableRegisterScope] interface,
// the 'postRegisterScopeHook' is registered
func (manager *ValueReferenceManager) registerHooks() {
	manager.source.RegisterPostSetHook(manager.postSetHook())

	if regSource, ok := manager.source.(HookablePostRegisterClass); ok {
		regSource.RegisterPostRegisterClassHook(manager.postRegisterClassHook())
	}
	if invSource, ok := manager.source.(HookablePostRegisterScope); ok {
		invSource.RegisterPostRegisterScopeHook(manager.postRegisterScopeHook())
	}
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
		return fmt.Errorf("failed to find references: %w", err)
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
	manager.allReferences = manager.allReferences[:len(manager.allReferences)-1]

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

// postSetHook is called after [reference.ValueTarget.SetPath] on the [ValueReferenceSource] is called.
// In order to ensure that the manager knows about all (new/removed) references,
// it will simply remove all known references at the path and add
// whatever references were introduced back.
func (manager *ValueReferenceManager) postSetHook() SetHookFunc {
	return func(path data.Path, value data.Value) error {
		newReferences, err := reference.FindValueReference(manager.source, reference.ValueReferenceRegex, path, value)
		if err != nil {
			return fmt.Errorf("failed to find references: %w", err)
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

// TODO: dont forget to register the class hooks on the new class(es)!

func (manager *ValueReferenceManager) postRegisterClassHook() RegisterClassHookFunc {
	return func(class *Class) error {
		// We cannot just use the class as source for the 'FindAllValueReferences' call.
		// This is because it will then make all reference target paths absolute to the class.
		// That's not what we want, the paths need to be relative to the managers source.
		// Hence we need to re-discover all references within the managers source.
		references, err := reference.FindAllValueReferences(manager.source)
		if err != nil {
			return err
		}

		// Add all references which originate from the newly added class.
		for _, ref := range references {
			if ref.Path.HasPrefix(class.Identifier) {
				err = manager.addReference(ref)
				if err != nil {
					return err
				}
			}
		}

		// make sure to also hook into the new class
		class.RegisterPostSetHook(manager.postSetHook())

		return nil
	}
}

func (manager *ValueReferenceManager) postRegisterScopeHook() RegisterScopeHookFunc {
	return func(scope Scope, registry *Registry) error {
		references, err := reference.FindAllValueReferences(manager.source)
		if err != nil {
			return err
		}

		for _, ref := range references {
			for _, class := range registry.ClassMap() {
				refPath := ref.Path.StripPrefix(data.NewPath(string(scope)))

				if refPath.HasPrefix(class.Identifier) {
					err = manager.addReference(ref)
					if err != nil {
						return err
					}
				}
			}
		}

		// register a postSet hook on the newly added registry, making sure that the scope prefix exists on all paths
		registry.RegisterPostSetHook(func(path data.Path, value data.Value) error {
			// make sure the path contains the scope
			path = path.Prepend(string(scope))
			return manager.postSetHook()(path, value)
		})

		return nil
	}
}

// ReferencesAtPath returns all known references at the given [data.Path]
func (manager *ValueReferenceManager) ReferencesAtPath(path data.Path) []reference.ValueReference {
	var refs []reference.ValueReference
	for _, ref := range manager.allReferences {
		if ref.Path.Equals(path) {
			refs = append(refs, ref)
		}
	}
	return refs
}

// AllReferences returns all known references. This slice may very well contain duplicates.
func (manager ValueReferenceManager) AllReferences() []reference.ValueReference {
	return manager.allReferences
}

// ReferenceMap returns a hash-map of all references.
// The key of the map is the result of the [reference.ValueReference.Hash] method.
func (manager ValueReferenceManager) ReferenceMap() map[string]reference.ValueReference {
	return manager.references
}
