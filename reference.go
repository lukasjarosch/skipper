package skipper

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/dominikbraun/graph"

	"github.com/lukasjarosch/skipper/data"
)

// TODO: PathReferences which allow yaml keys to be references as well
// TODO: test with Registry
// TODO: test with Inventory

var (
	// ReferenceRegex defines the strings which are valid references
	// See: https://regex101.com/r/lIuuep/1
	ReferenceRegex = regexp.MustCompile(`\${(?P<reference>[\w-]+(?:\:[\w-]+)*)}`)

	ErrUndefinedReferenceTarget = fmt.Errorf("undefined reference target path")
	ErrReferenceSourceIsNil     = fmt.Errorf("reference source is nil")
	ErrSelfReferencingReference = fmt.Errorf("self-referencing reference")
	ErrCyclicReference          = fmt.Errorf("cyclic reference")
)

// ValueReference is a reference to a value with a different path.
type ValueReference struct {
	// Path is the path where the reference is defined
	Path data.Path
	// TargetPath is the path how it was written within the source.
	TargetPath data.Path
	// AbsoluteTargetPath is the TargetPath, but absolute to the source
	AbsoluteTargetPath data.Path
	// If the value to which the reference (AbsoluteTargetPath) points to
	// also contains references, then these are added as ChildReferences.
	ChildReferences []ValueReference
}

func (ref ValueReference) Name() string {
	return fmt.Sprintf("${%s}", strings.ReplaceAll(ref.TargetPath.String(), ".", ":"))
}

type ReferenceValueSource interface {
	DataGetter
	AbsolutePathMaker
	// Values must return a map of absolute paths to their respective values.
	// If the paths are not absolute then working with references will cause
	// all sort of problems.
	Values() map[string]data.Value
}

// ParseValueReferences will, given the [ReferenceValueSource] extract a list
// of [ValueReference]s which can then be resolved and ultimately replaced.
func ParseValueReferences(source ReferenceValueSource) ([]ValueReference, error) {
	if source == nil {
		return nil, ErrReferenceSourceIsNil
	}

	// Discover all references within the source.
	var references []ValueReference
	for path, value := range source.Values() {
		refs, err := FindValueReferences(source, data.NewPath(path), value)
		if err != nil {
			return nil, err
		}
		references = append(references, refs...)
	}

	// Now, for each reference we need to determine whether their targetValue contains
	// references again. If it does, add them as ChildReferences.
	for _, ref := range references {
		val, err := source.GetPath(ref.AbsoluteTargetPath)
		if err != nil {
			return nil, err
		}

		childReferences, err := FindValueReferences(source, ref.AbsoluteTargetPath, val)
		if err != nil {
			return nil, err
		}

		ref.ChildReferences = append(ref.ChildReferences, childReferences...)
		spew.Dump(val)
	}

	return references, nil
}

func FindValueReferences(source AbsolutePathMaker, path data.Path, value data.Value) ([]ValueReference, error) {
	var references []ValueReference
	referenceMatches := ReferenceRegex.FindAllStringSubmatch(value.String(), -1)
	if referenceMatches != nil {
		for _, match := range referenceMatches {
			// References can be relative to a Class. But that is hard to work with within a Registry or Inventory
			// as a class-relative path can be valid within multiple classes / scopes.
			// Hence an absolute path is resolved which we will be working with from now on.
			// This works because the path returned by the 'Values()' call is expected to be
			// absolute already. This then defines the context in which the reference is defined
			// which in turn is used to resolve the absolute path to which the reference points to.
			absPath, err := source.AbsolutePath(ReferencePathToPath(match[1]), path)
			if err != nil {
				return nil, fmt.Errorf("unable to resolve absolute path of '%s': %w", match[1], err)
			}

			references = append(references, ValueReference{
				Path:               path,
				TargetPath:         ReferencePathToPath(match[1]),
				AbsoluteTargetPath: absPath,
			})
		}
	}
	return references, nil
}

type OldParseSource interface {
	DataWalker
	AbsolutePathMaker
}

// ParseReferences will use the [ReferenceSourceWalker] to traverse all values
// and search for References within those values.
// The returned slice of references contains all found references, even
// duplicates if a reference is used multiple times.
func ParseReferences(source OldParseSource) ([]ValueReference, error) {
	if source == nil {
		return nil, ErrReferenceSourceIsNil
	}

	var references []ValueReference
	source.Walk(func(path data.Path, value data.Value, isLeaf bool) error {
		if !isLeaf {
			return nil
		}
		referenceMatches := ReferenceRegex.FindAllStringSubmatch(value.String(), -1)

		if referenceMatches != nil {
			for _, match := range referenceMatches {
				absPath, err := source.AbsolutePath(ReferencePathToPath(match[1]), path)
				if err != nil {
					return fmt.Errorf("unable to resolve absolute path of '%s': %w", match[1], err)
				}
				ref := ValueReference{
					Path:               path,
					TargetPath:         ReferencePathToPath(match[1]),
					AbsoluteTargetPath: absPath,
				}

				references = append(references, ref)
			}
		}

		return nil
	})

	return references, nil
}

// ReferencesInValue returns all references within the passed value.
// Note that the returned references do not have a 'Path' set!
// TODO: replace
func ReferencesInValue(value data.Value) []ValueReference {
	var references []ValueReference

	referenceMatches := ReferenceRegex.FindAllStringSubmatch(value.String(), -1)
	if referenceMatches != nil {
		for _, match := range referenceMatches {
			references = append(references, ValueReference{
				TargetPath: ReferencePathToPath(match[1]),
				Path:       data.Path{},
			})
		}
	}

	return references
}

type referenceVertex struct {
	ValueReference
	RawValue data.Value
}

type ReferenceSource interface {
	DataSetterGetter
	DataWalker
}

type ReferenceSourceRelativeGetter interface {
	DataGetter
	GetClassRelativePath(data.Path, data.Path) (data.Value, error)
}

// ResolveReferences will resolve dependencies between all given references and return
// a sorted slice of the same references. This represents the order in which references should
// be replaced without causing dependency issues.
//
// Note that the list of references passed must contain all existing references, even duplicates.
func ResolveReferences(references []ValueReference, resolveSource DataGetter) ([]ValueReference, error) {
	if resolveSource == nil {
		return nil, ErrReferenceSourceIsNil
	}

	// used by the graph to tell vertecies apart
	referenceVertexHash := func(ref referenceVertex) string {
		return ref.TargetPath.String()
	}

	g := graph.New(referenceVertexHash, graph.Acyclic(), graph.Directed(), graph.PreventCycles())

	// Register all references as vertecies
	var vertecies []referenceVertex
	for _, ref := range references {
		rawValue, err := resolveSource.GetPath(ref.TargetPath)
		if err != nil {
			// In case the path cannot be resolved it might be a class-local reference.
			// We can attempt to resolve the Class since we know where the reference is located.
			if errors.Is(err, ErrPathNotFound) {
				if relativeSource, ok := resolveSource.(ReferenceSourceRelativeGetter); ok {
					rawValue, err = relativeSource.GetClassRelativePath(ref.Path, ref.TargetPath)
					if err != nil {
						return nil, fmt.Errorf("%w: %s at path '%s': %w", ErrUndefinedReferenceTarget, ref.Name(), ref.Path, err)
					}
				}
				return nil, fmt.Errorf("%w: %s at path '%s': %w", ErrUndefinedReferenceTarget, ref.Name(), ref.Path, err)
			} else {
				return nil, fmt.Errorf("%w: %s at path '%s': %w", ErrUndefinedReferenceTarget, ref.Name(), ref.Path, err)
			}
		}
		node := referenceVertex{ValueReference: ref, RawValue: rawValue}
		err = g.AddVertex(node)
		if err != nil {
			// References can occur multiple times, but we only need to resolve them once.
			// So we can ignore the ErrVertexAlreadyExists error.
			if !errors.Is(err, graph.ErrVertexAlreadyExists) {
				return nil, err
			}
		}
		vertecies = append(vertecies, node)
	}

	// Create edges between dependent references by looking at the actual value the 'TargetPath' points to.
	// If that value contains more references, then these are dependencies of the currently examined node (reference).
	for _, referenceVertex := range vertecies {
		referenceDependencies := ReferencesInValue(referenceVertex.RawValue)

		for _, referenceDependency := range referenceDependencies {
			// The dependency of the current vertex must already be a vertex in the graph; fetch it.
			n, err := g.Vertex(referenceDependency.TargetPath.String())
			if err != nil {
				return nil, fmt.Errorf("%s: %w", referenceDependency.Name(), err)
			}

			if referenceVertex.TargetPath.Equals(n.TargetPath) {
				return nil, fmt.Errorf("%s: %w", referenceVertex.Name(), ErrSelfReferencingReference)
			}

			err = g.AddEdge(referenceVertex.TargetPath.String(), n.TargetPath.String())
			if err != nil {
				if errors.Is(err, graph.ErrEdgeCreatesCycle) {
					return nil, fmt.Errorf("%s -> %s: %w", referenceVertex.Name(), referenceDependency.Name(), ErrCyclicReference)
				}
				// If the edge already exists, we do not need to add it again. Hence we ignore that error.
				if !errors.Is(err, graph.ErrEdgeAlreadyExists) {
					return nil, err
				}
			}
		}
	}

	// Perform a stable topological sort of the dependency graph.
	// The returned order is stable in that the references are sorted
	// by their name length or alphabetically if they are the same length.
	// This eliminates the issue that the actual topological sorting algorithm usually
	// has multiple valid solutions.
	order, err := graph.StableTopologicalSort[string, referenceVertex](g, func(s1, s2 string) bool {
		// Strings are of different length, sort by length
		if len(s1) != len(s2) {
			return len(s1) < len(s2)
		}
		// Otherwise, sort alphabetically
		return s1 > s2
	})
	if err != nil {
		return nil, fmt.Errorf("failed to sort reference graph: %w", err)
	}

	// Perform TopologicalSort which returns a list of strings (TargetPaths)
	// starting with the one which has the most dependencies.
	// order, err := graph.TopologicalSort[string, referenceVertex](g)
	// if err != nil {
	// 	return nil, err
	// }

	// We need to reverse the result from the TopologicalSort.
	// This is because the reference without dependencies will be sorted 'at the end'.
	// But we want to resolve them first.
	var reversedOrder []data.Path
	for i := len(order) - 1; i >= 0; i-- {
		reversedOrder = append(reversedOrder, data.NewPath(order[i]))
	}

	// Now that we have the order in which references must be replaced,
	// lets finally re-order the passed (non-deduplicated) input references.
	var orderedReferences []ValueReference
	for _, refOrder := range reversedOrder {
		for _, ref := range references {
			if ref.TargetPath.Equals(refOrder) {
				orderedReferences = append(orderedReferences, ref)
			}
		}
	}

	// sanity check
	if len(orderedReferences) != len(references) {
		return nil, fmt.Errorf("unexpected amount of resolved references, this is a bug")
	}

	return orderedReferences, nil
}

type ReferenceSourceSetter interface {
	SetPath(data.Path, interface{}) error
}

func ReplaceReferences(references []ValueReference, source DataSetterGetter) error {
	if source == nil {
		return ErrReferenceSourceIsNil
	}

	for _, reference := range references {
		targetValue, err := source.GetPath(reference.TargetPath)
		if err != nil {
			// In case the path cannot be resolved it might be a class-local reference.
			// We can attempt to resolve the Class since we know where the reference is located.
			if errors.Is(err, ErrPathNotFound) {
				if relativeSource, ok := source.(ReferenceSourceRelativeGetter); ok {
					targetValue, err = relativeSource.GetClassRelativePath(reference.Path, reference.TargetPath)
					if err != nil {
						return fmt.Errorf("cannot resolve reference target path: %w", err)
					}
				}
			} else {
				return fmt.Errorf("cannot resolve reference target path: %w", err)
			}
		}

		sourceValue, err := source.GetPath(reference.Path)
		if err != nil {
			return fmt.Errorf("cannot resolve reference path: %w", err)
		}

		// If the sourceValue only contains the reference, then
		// we just use the 'SetPath' function in order to preserve the type of targetValue.
		// This is required to allow replacing maps and arrays.
		if len(sourceValue.String()) == len(reference.Name()) {
			err = source.SetPath(reference.Path, targetValue.Raw)
			if err != nil {
				return err
			}
			continue
		}

		// In this case the reference is 'embedded', e.g. "Hello there ${name}",
		// therefore we can only perform a string replacement to not erase the surrounding context.
		replacedValue := strings.Replace(sourceValue.String(), reference.Name(), targetValue.String(), 1)
		err = source.SetPath(reference.Path, replacedValue)
		if err != nil {
			return err
		}
	}

	return nil
}

// ReferencePathToPath converts the path used within references (colon-separated) to a proper [data.Path]
func ReferencePathToPath(referencePath string) data.Path {
	referencePath = strings.ReplaceAll(referencePath, ":", ".")
	return data.NewPath(referencePath)
}
