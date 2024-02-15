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

type ResolvedReference struct {
	Reference
	// TargetValue is the value to which the TargetPath points to.
	// If TargetReference is not nil, this value must be [data.NilValue].
	TargetValue data.Value
	// TargetReference is non-nil if the Reference points to another [ResolvedReference]
	// If the Reference just points to a scalar value, this will be nil.
	TargetReference *ResolvedReference
}

type ReferenceSourceWalker interface {
	WalkValues(func(path data.Path, value data.Value) error) error
}

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
			return nil, fmt.Errorf("%w: %s at path %s", ErrUndefinedReferenceTarget, ref.Name(), ref.Path)
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

// ReferencePathToPath converts the path used within references (colon-separated) to a proper [data.Path]
func ReferencePathToPath(referencePath string) data.Path {
	referencePath = strings.ReplaceAll(referencePath, ":", ".")
	return data.NewPath(referencePath)
}
