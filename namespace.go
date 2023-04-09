package skipper

import (
	"errors"
	"strings"
)

var (
	ErrEmptyNamespace = errors.New("namespace is empty")
)

const NamespaceSeparator = "."

type Namespace string

func NewNamespace(ns string) (Namespace, error) {
	if len(ns) == 0 {
		return "", ErrEmptyNamespace
	}
	return Namespace(ns), nil
}

func (ns Namespace) Segments() []string {
	return strings.Split(ns.String(), NamespaceSeparator)
}

func (ns Namespace) String() string {
	return string(ns)
}
