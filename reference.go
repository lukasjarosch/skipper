package skipper

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

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
	ErrReferenceCycle           = fmt.Errorf("reference cycles are not allowed")
)

// Reference is a reference to a value with a different path.
type Reference struct {
	// Path is the path where the reference is defined
	Path data.Path
	// TargetPath is the path the reference points to
	TargetPath data.Path
}

func (ref Reference) Name() string {
	return fmt.Sprintf("${%s}", strings.ReplaceAll(ref.TargetPath.String(), ".", ":"))
}

type ReferenceSourceWalker interface {
	WalkValues(func(path data.Path, value data.Value) error) error
}

// ParseReferences will use the [ReferenceSourceWalker] to traverse all values
// and search for References within those values.
// The returned slice of references contains all found references, even
// duplicates if a reference is used multiple times.
func ParseReferences(source ReferenceSourceWalker) ([]Reference, error) {
	if source == nil {
		return nil, ErrReferenceSourceIsNil
	}

	var references []Reference
	source.WalkValues(func(path data.Path, value data.Value) error {
		referenceMatches := ReferenceRegex.FindAllStringSubmatch(value.String(), -1)

		if referenceMatches != nil {
			for _, match := range referenceMatches {
				references = append(references, Reference{
					Path:       path,
					TargetPath: ReferencePathToPath(match[1]),
				})
			}
		}

		return nil
	})

	return references, nil
}

// ReferencesInValue returns all references within the passed value.
// Note that the returned references do not have a 'Path' set!
func ReferencesInValue(value data.Value) []Reference {
	var references []Reference

	referenceMatches := ReferenceRegex.FindAllStringSubmatch(value.String(), -1)
	if referenceMatches != nil {
		for _, match := range referenceMatches {
			references = append(references, Reference{
				TargetPath: ReferencePathToPath(match[1]),
				Path:       data.Path{},
			})
		}
	}

	return references
}

type referenceVertex struct {
	Reference
	RawValue data.Value
}

type ReferenceSourceGetter interface {
	GetPath(path data.Path) (data.Value, error)
}

type ReferenceSourceRelativeGetter interface {
	GetClassRelativePath(data.Path, data.Path) (data.Value, error)
}

// ResolveReferences will resolve dependencies between all given references and return
// a sorted slice of the same references. This represents the order in which references should
// be replaced without causing dependency issues.
//
// Note that the list of references passed must contain all existing references, even duplicates.
func ResolveReferences(references []Reference, resolveSource ReferenceSourceGetter) ([]Reference, error) {
	referenceNodeHash := func(ref referenceVertex) string {
		return ref.TargetPath.String()
	}

	g := graph.New(referenceNodeHash, graph.Acyclic(), graph.Directed(), graph.PreventCycles())

	// Register all references as nodes
	var nodes []referenceVertex
	for _, ref := range references {
		rawValue, err := resolveSource.GetPath(ref.TargetPath)
		if err != nil {
			// In case the path cannot be resolved it might be a class-local reference.
			// We can attempt to resolve the Class since we know where the reference is located.
			if errors.Is(err, ErrPathNotFound) {
				if relativeSource, ok := resolveSource.(ReferenceSourceRelativeGetter); ok {
					rawValue, err = relativeSource.GetClassRelativePath(ref.Path, ref.TargetPath)
					if err != nil {
						return nil, fmt.Errorf("%w: %s at path %s: %w", ErrUndefinedReferenceTarget, ref.Name(), ref.Path, err)
					}
				}
			} else {
				return nil, fmt.Errorf("%w: %s at path %s: %w", ErrUndefinedReferenceTarget, ref.Name(), ref.Path, err)
			}
		}
		node := referenceVertex{Reference: ref, RawValue: rawValue}
		err = g.AddVertex(node)
		if err != nil {
			// References can occur multiple times, but we only need to resolve them once.
			// So we can ignore the ErrVertexAlreadyExists error.
			if !errors.Is(err, graph.ErrVertexAlreadyExists) {
				return nil, err
			}
		}
		nodes = append(nodes, node)
	}

	// Create edges between dependent references by looking at the actual value the 'TargetPath' points to.
	// If that value contains more references, then these are dependencies of the currently examined node (reference).
	for _, refNode := range nodes {
		referenceDependencies := ReferencesInValue(refNode.RawValue)

		for _, refDep := range referenceDependencies {
			n, err := g.Vertex(refDep.TargetPath.String())
			if err != nil {
				return nil, err
			}

			if refNode.TargetPath.Equals(n.TargetPath) {
				return nil, fmt.Errorf("self-referencing reference %s at path %s", refNode.Name(), refNode.Path)
			}

			err = g.AddEdge(refNode.TargetPath.String(), n.TargetPath.String())
			if err != nil {
				if errors.Is(err, graph.ErrEdgeCreatesCycle) {
					return nil, fmt.Errorf("reference %s -> %s would introduce cycle: %w", refNode.Name(), refDep.Name(), ErrReferenceCycle)
				}
				return nil, err
			}
		}
	}

	// Perform TopologicalSort which returns a list of strings (TargetPaths)
	// starting with the one which has the most dependencies.
	order, err := graph.TopologicalSort[string, referenceVertex](g)
	if err != nil {
		return nil, err
	}

	// We need to reverse the result from the TopologicalSort.
	// This is because the reference without dependencies will be sorted 'at the end'.
	// But we want to resolve them first.
	var reversedOrder []data.Path
	for i := len(order) - 1; i >= 0; i-- {
		reversedOrder = append(reversedOrder, data.NewPath(order[i]))
	}

	// Now that we have the order in which references must be replaced,
	// lets finally re-order the passed (non-deduplicated) input references.
	var orderedReferences []Reference
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

type ReferenceSourceGetterSetter interface {
	ReferenceSourceGetter
	ReferenceSourceSetter
}

func ReplaceReferences(references []Reference, source ReferenceSourceGetterSetter) error {
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
