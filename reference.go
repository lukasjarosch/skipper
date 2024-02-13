package skipper

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/lukasjarosch/skipper/data"
)

// TODO: handle reference-to-reference
// TODO: PathReferences / KeyReferences
// TODO: handle cyclic references

var (
	// ReferenceRegex defines the strings which are valid references
	// See: https://regex101.com/r/lIuuep/1
	ReferenceRegex = regexp.MustCompile(`\${(?P<reference>[\w-]+(?:\:[\w-]+)*)}`)

	ErrUndefinedReference = fmt.Errorf("undefined reference")
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
	// TargetValue is the value to which the TargetPath points to
	// This can be [data.NilValue]. In that case there is no target
	// value but a target reference.
	TargetValue data.Value
	// TargetReference is non-nil if the Reference points to another [Reference]
	TargetReference *Reference
}

// ReferenceParser is responsible for discovering and resolving references.
type ReferenceParser struct {
	source ReferenceSourceWalker
}

type ReferenceSourceWalker interface {
	WalkValues(func(path data.Path, value data.Value) error) error
}

type ReferenceSourceGetter interface {
	GetPath(path data.Path) (data.Value, error)
}

var ErrReferenceSourceIsNil = fmt.Errorf("source is nil")

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

func ResolveReferences(references []Reference, resolveSource ReferenceSourceGetter) ([]ResolvedReference, error) {
	if resolveSource == nil {
		return nil, ErrReferenceSourceIsNil
	}

	var errs error
	var resolvedReferences []ResolvedReference
	for _, reference := range references {
		val, err := resolveSource.GetPath(reference.TargetPath)
		if err != nil {
			errs = errors.Join(errs, fmt.Errorf("%w %s at %s: %w", ErrUndefinedReference, reference.Name(), reference.Path, err))
			continue
		}

		resolvedReferences = append(resolvedReferences, ResolvedReference{
			Reference:   reference,
			TargetValue: val,
		})
	}
	if errs != nil {
		return nil, errs
	}

	return resolvedReferences, nil
}

// ReferencePathToPath converts the path used within references (colon-separated) to a proper [data.Path]
func ReferencePathToPath(referencePath string) data.Path {
	referencePath = strings.ReplaceAll(referencePath, ":", ".")
	return data.NewPath(referencePath)
}
