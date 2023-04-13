package skipper

import (
	"os"
	"path/filepath"
	"strings"
)

// FilePathToPath converts a file path from a filesystem to a skipper [Path] by
// - removing the file extension
// - replacing path separators with [skipper.PathSeparator]
// - removing the commonPrefix if it is not an empty string
func FilePathToPath(filePath string, commonPrefix string) Path {
	path := filePath[:len(filePath)-len(filepath.Ext(filePath))]
	path = strings.ReplaceAll(path, string(os.PathSeparator), PathSeparator)

	if len(commonPrefix) > 0 {
		path = strings.Join(strings.Split(path, PathSeparator)[1:], PathSeparator)
	}

	return P(path)
}
