package graph

import (
	"os"

	"github.com/dominikbraun/graph"
	"github.com/dominikbraun/graph/draw"
)

// visualize is mainly a debugging function which will render a DOT file
// of the graph into the given filePath and add the label as graph description.
func visualize[K comparable, T any](graph graph.Graph[K, T], filePath string, label string) error {
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
