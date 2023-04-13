package skipper

import (
	"os"
	"path/filepath"
	"strings"
)

// FilePathToPath converts a file path from a filesystem to a skipper [Path] by
// - removing any leading or trailing path separators in both args
// - removing the commonPrefix from the filePath
// - removing the file extension
// - replacing path separators with [skipper.PathSeparator]
func FilePathToPath(filePath string, commonPrefix string) Path {
	commonPrefix = strings.Trim(commonPrefix, string(os.PathSeparator))
	path := strings.Trim(filePath, string(os.PathSeparator))

	path = strings.TrimLeft(path, commonPrefix)
	path = path[:len(path)-len(filepath.Ext(path))]
	path = strings.ReplaceAll(path, string(os.PathSeparator), PathSeparator)

	return P(path)
}
