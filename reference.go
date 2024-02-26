package skipper

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/dominikbraun/graph"
	"github.com/dominikbraun/graph/draw"

	"github.com/lukasjarosch/skipper/data"
)

// TODO: PathReferences which allow yaml keys to be references as well
// TODO: test with Registry
// TODO: test with Inventory

var (
	// ValueReferenceRegex defines the strings which are valid references
	// See: https://regex101.com/r/lIuuep/1
	ValueReferenceRegex = regexp.MustCompile(`\${(?P<reference>[\w-]+(?:\:[\w-]+)*)}`)

	ErrUndefinedReferenceTarget = fmt.Errorf("undefined reference target path")
	ErrReferenceSourceIsNil     = fmt.Errorf("reference source is nil")
	ErrReferenceTargetIsNil     = fmt.Errorf("reference target is nil")
	ErrSelfReferencingReference = fmt.Errorf("self-referencing reference")
	ErrCyclicReference          = fmt.Errorf("cyclic reference")
)

type ValueReferenceSource interface {
	AbsolutePathMaker
	// Values must return a map of absolute (!) paths to their respective [data.Value]s.
	Values() map[string]data.Value
}

// ValueReference is a reference to a value with a different path.
type ValueReference struct {
	// Path is the path where the reference is defined
	Path data.Path
	// TargetPath is the path how it was written within the source.
	TargetPath data.Path
	// AbsoluteTargetPath is the TargetPath, but absolute to the source
	AbsoluteTargetPath data.Path
}

// Name returns the full name of the reference, just as it was written within the data.
// Note that this DOES NOT use the AbsoluteTargetPath on purpose!
// That path is calculated and will
//  1. confuse anyone if the reference suddenly uses a calculated full path
//  2. replacing the reference will not work because there we need the exact string how it appeared in the data.
func (ref ValueReference) Name() string {
	return fmt.Sprintf("${%s}", strings.ReplaceAll(ref.TargetPath.String(), ".", ":"))
}

// Hash returns the reference-unique hash which is used to de-duplicate references.
func (ref ValueReference) Hash() string {
	return fmt.Sprintf("%s|%s", ref.Path, ref.AbsoluteTargetPath)
}

// ValueReferenceManager is responsible for handling [ValueReference]s.
// It can parse, resolve and replace any value references.
type ValueReferenceManager struct {
	source ValueReferenceSource
	// allReferences contains all found references, even duplicates
	allReferences []ValueReference
	// references maps reference hashes to references, hence being
	// the deduplicated version of allReferences
	references map[string]ValueReference
	// stores all references and their dependencies
	dependencyGraph graph.Graph[string, ValueReference]
}

// TODO: needs to provide function which is hooked into PostSetPath in each class
// This is because references need to be re-evaluated if anything changes
func NewValueReferenceManager(source ValueReferenceSource) (*ValueReferenceManager, error) {
	if source == nil {
		return nil, ErrReferenceSourceIsNil
	}

	m := &ValueReferenceManager{
		source: source,
	}

	references, err := FindValueReferences(m.source)
	if err != nil {
		return nil, err
	}
	m.allReferences = references

	// deduplicate references in map
	for _, ref := range m.allReferences {
		m.references[ref.Hash()] = ref
	}

	return m, nil
}

// ValueReferenceTarget is the required interface in order to replace references.
type ValueReferenceTarget interface {
	DataGetter
	DataSetter
}

// ReplaceValueReferences will replace all given references within the given ValueReferenceTarget.
func ReplaceValueReferences(target ValueReferenceTarget, references []ValueReference) error {
	if references == nil || len(references) == 0 {
		return nil
	}
	if target == nil {
		return ErrReferenceTargetIsNil
	}

	for _, reference := range references {
		// The targetValue is the value to which the reference points to.
		targetValue, err := target.GetPath(reference.AbsoluteTargetPath)
		if err != nil {
			return err
		}

		// The sourceValue is the value in which the reference is found and where it is
		// going to be replaced by the targetValue
		sourceValue, err := target.GetPath(reference.Path)
		if err != nil {
			return err
		}

		// If the sourceValue contains exactly the reference and nothing else,
		// we can just swap out the raw value completely.
		if strings.EqualFold(sourceValue.String(), reference.Name()) {
			err = target.SetPath(reference.Path, targetValue.Raw)
			if err != nil {
				return err
			}
			continue
		}

		// If the reference is embedded within literals (e.g. 'hello there ${name}'),
		// then we need to just perform one string replacement the sourceValue.
		// We do this only once, even if the same reference may exist multiple times.
		// But since the references slice can contain duplicates, we expect the
		// references to appear multiple times in the slice in such cases.
		replacedValue := strings.Replace(sourceValue.String(), reference.Name(), targetValue.String(), 1)
		err = target.SetPath(reference.Path, replacedValue)
		if err != nil {
			return err
		}
	}

	return nil
}

// ReorderValueReferences will take a slice of ordered, deduplicated ValueReferences
// and a slice of unordered, non-deduplicated ValueReferences.
// The unordered slice is then rearranged based and returned as determined by the order.
// References are compared using their `Hash` function.
// If the order is empty/nil, then allReferences is just re-emitted.
// If allReferences is empty/nil, then nil is returned.
func ReorderValueReferences(order []ValueReference, allReferences []ValueReference) []ValueReference {
	if order == nil || len(order) == 0 {
		return allReferences
	}
	if allReferences == nil || len(allReferences) == 0 {
		return nil
	}

	var orderedReferences []ValueReference
	for _, orderedReference := range order {
		for _, ref := range allReferences {
			if orderedReference.Hash() == ref.Hash() {
				orderedReferences = append(orderedReferences, ref)
			}
		}
	}
	return orderedReferences
}

// CalculateReplacementOrder takes an dependencyGraph and performs a stable topological sort.
// The returned slice of ValueReferences is in the order in which the references
// can be replaced without any dependency issues like re-introducing references during replacement.
func CalculateReplacementOrder(dependencyGraph graph.Graph[string, ValueReference]) ([]ValueReference, error) {
	// Perform a stable topological sort of the dependency graph.
	// The returned orderedHashes is stable in that the references are sorted
	// by their name length or alphabetically if they are the same length.
	// This eliminates the issue that the actual topological sorting algorithm usually
	// has multiple valid solutions.
	orderedHashes, err := graph.StableTopologicalSort[string, ValueReference](dependencyGraph, func(s1, s2 string) bool {
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

	// The result of the topological sorting is reverse to what we want.
	// We expect the ValueReference without any dependencies to be at the top.
	var reversedOrder []ValueReference
	for i := len(orderedHashes) - 1; i >= 0; i-- {
		ref, err := dependencyGraph.Vertex(orderedHashes[i])
		if err != nil {
			return nil, err
		}
		reversedOrder = append(reversedOrder, ref)
	}

	return reversedOrder, nil
}

// BuildDependencyGraph will take a list of ValueReferences, deduplicate it and then create
// a vertex within a graph for each ValueReference.
// Then, for each ValueReference, all dependent ValueReference (i.e. if the reference points to a value which again contains a reference)
// are resolved and directed edges are created.
// The parameters of the graph ensure that the resulting dependencyGraph is acyclic, directed and does not contain any self-references.
func BuildDependencyGraph(references []ValueReference) (graph.Graph[string, ValueReference], error) {
	vertexHashFunc := func(ref ValueReference) string {
		return ref.Hash()
	}

	references = DeduplicateValueReferences(references)

	dependencyGraph := graph.New(vertexHashFunc, graph.Acyclic(), graph.Directed(), graph.PreventCycles())

	// Tegister all references as graph vertecies
	for _, reference := range references {
		err := dependencyGraph.AddVertex(reference)
		if err != nil {
			return nil, fmt.Errorf("cannot add reference %s: %w", reference.Name(), err)
		}
	}

	// Create edges between dependent references
	for _, reference := range references {
		dependencies := ResolveDependantValueReferences(reference, references)

		for _, dependency := range dependencies {
			dependencyVertex, err := dependencyGraph.Vertex(dependency.Hash())
			if err != nil {
				return nil, fmt.Errorf("unexpectedly could not fetch reference vertex %s: %w", dependency.Hash(), err)
			}

			// prevent self-referencing references
			if dependencyVertex.AbsoluteTargetPath.Equals(reference.AbsoluteTargetPath) {
				return nil, fmt.Errorf("%s: %w", reference.Name(), ErrSelfReferencingReference)
			}

			err = dependencyGraph.AddEdge(reference.Hash(), dependency.Hash())
			if err != nil {
				if errors.Is(err, graph.ErrEdgeCreatesCycle) {
					return nil, fmt.Errorf("%s -> %s: %w", reference.Name(), dependency.Name(), ErrCyclicReference)
				}
				return nil, fmt.Errorf("failed to register dependency %s: %w", dependency.Name(), err)
			}
		}
	}

	return dependencyGraph, nil
}

// VisualizeDependencyGraph is mainly a debugging function which will render a DOT file
// of the dependencyGraph into the given filePath and add the label as graph description.
func VisualizeDependencyGraph(graph graph.Graph[string, ValueReference], filePath string, label string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	err = draw.DOT(graph, file,
		draw.GraphAttribute("label", label),
		draw.GraphAttribute("overlap", "prism"))
	if err != nil {
		return err
	}

	return nil
}

// DeduplicateValueReferences takes a list of ValueReferences and
// deduplicates them based on the `Hash`.
func DeduplicateValueReferences(references []ValueReference) []ValueReference {
	visited := make(map[string]bool)
	deduplicated := []ValueReference{}

	for _, ref := range references {
		if _, seen := visited[ref.Hash()]; !seen {
			visited[ref.Hash()] = true
			deduplicated = append(deduplicated, ref)
		}
	}

	return deduplicated
}

// ResolveDependantValueReferences takes a reference and a list of all existing references
// and checks whether the AbsoluteTargetPath of said reference is the Path of any other reference.
// Note that even if allReferences may not be deduplicated, the function will only work with a deduplicated version.
// So if a reference `${person:name}` exists, as well as a reference `${common:name}` at path `person.name`,
// then the latter is a direct dependency of the `${person:name}` reference.
func ResolveDependantValueReferences(reference ValueReference, allReferences []ValueReference) []ValueReference {
	var dependencies []ValueReference

	// If the reference's AbsoluteTargetPath is a Path of any existing reference,
	// the references depend on each other.
	for _, ref := range DeduplicateValueReferences(allReferences) {
		if reference.AbsoluteTargetPath.Equals(ref.Path) {
			dependencies = append(dependencies, ref)
		}
	}

	return dependencies
}

// FindValueReferences searches for all ValueReferences within the given ValueReferenceSource.
func FindValueReferences(source ValueReferenceSource) ([]ValueReference, error) {
	var references []ValueReference
	var errs error

	for path, value := range source.Values() {
		referenceTargetPaths := FindReferenceTargetPaths(ValueReferenceRegex, value)

		// Create ValueReference structs for every found referenceTargetPaths
		for _, refTargetPath := range referenceTargetPaths {
			absTargetPath, err := source.AbsolutePath(refTargetPath, data.NewPath(path))
			if err != nil {
				return nil, errors.Join(errs, err)
			}

			ref := ValueReference{
				Path:               data.NewPath(path),
				TargetPath:         refTargetPath,
				AbsoluteTargetPath: absTargetPath,
			}
			references = append(references, ref)
		}
	}

	if errs != nil {
		return nil, errs
	}

	return references, nil
}

// FindReferenceTargetPaths yields all the targetPaths of any references contained within the value.
// If the value is 'foo ${bar:baz}', then the returned path would be `bar.baz`.
// The returned paths should be considered relative!
func FindReferenceTargetPaths(regex *regexp.Regexp, value data.Value) []data.Path {
	var targetPaths []data.Path

	referenceMatches := regex.FindAllStringSubmatch(value.String(), -1)
	if referenceMatches == nil {
		return nil
	}

	for _, match := range referenceMatches {
		// If the regex itself is malformed, we can do nothing but panic.
		// The first element (match[0]) will be the full string (aka the input).
		// And the second element is expected to contain the part between the brackets.
		if len(match) < 2 {
			panic("regex match has not enough elements; this is a bug in the regex itself!")
		}
		targetPaths = append(targetPaths, ReferencePathToPath(match[1]))
	}

	return targetPaths
}

// ReferencePathToPath converts the path used within references (colon-separated) to a proper [data.Path]
func ReferencePathToPath(referencePath string) data.Path {
	referencePath = strings.ReplaceAll(referencePath, ":", ".")
	return data.NewPath(referencePath)
}
